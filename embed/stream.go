package embed

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// StreamAsset is an asset whose data is read from Reader. RenderTo consumes the
// reader once and spools its contents to temporary storage, so even non-seekable
// sources such as stdin need only bounded memory.
type StreamAsset struct {
	Name, Parent string
	Reader       io.Reader
}

type spooledAsset struct {
	file   *os.File
	size   int64
	marker []byte
	token  string
	digest string
}

// RenderTo generates source without retaining input or generated data in memory.
// Temporary disk space equal to the total input size is required. Target language
// compilers and runtimes may impose much smaller limits than mmmbed itself.
// On error, w may already contain partial output.
func RenderTo(w io.Writer, language string, inputs []StreamAsset, options Options) error {
	if len(inputs) == 0 {
		return fmt.Errorf("no assets")
	}
	if len(options.Variables) != 0 && len(options.Variables) != len(inputs) {
		return fmt.Errorf("provide one variable per input")
	}
	lang := strings.ToLower(language)
	files := make([]spooledAsset, 0, len(inputs))
	defer func() {
		for _, a := range files {
			a.file.Close()
			os.Remove(a.file.Name())
		}
	}()
	assets := make([]Asset, 0, len(inputs))
	for _, a := range inputs {
		if a.Reader == nil {
			return fmt.Errorf("nil reader for %q", a.Name)
		}
		f, err := os.CreateTemp("", "mmmbed-asset-*")
		if err != nil {
			return err
		}
		files = append(files, spooledAsset{file: f})
		index := len(files) - 1
		hash := sha256.New()
		size, err := io.Copy(io.MultiWriter(f, hash), a.Reader)
		if err != nil {
			return fmt.Errorf("read %q: %w", a.Name, err)
		}
		files[index].size = size
		files[index].digest = hex.EncodeToString(hash.Sum(nil)[:3])
		marker := make([]byte, 32)
		if _, err = rand.Read(marker); err != nil {
			return err
		}
		files[index].marker = marker
		assets = append(assets, Asset{Name: a.Name, Parent: a.Parent, Data: marker})
	}
	if len(options.Variables) == 0 {
		used := map[string]bool{}
		options.names = make([]string, len(inputs))
		for i, a := range files {
			base := "mmmbed_file_" + a.digest
			name := base
			for n := 2; used[name]; n++ {
				name = fmt.Sprintf("%s_%d", base, n)
			}
			used[name] = true
			options.names[i] = name
		}
	}
	template, err := RenderWithOptions(lang, assets, options)
	if err != nil {
		return err
	}
	if lang == "c" || lang == "cpp" {
		for i, a := range files {
			name := options.names
			if name == nil {
				name = options.Variables
			}
			template = []byte(strings.Replace(string(template), fmt.Sprintf("{(const char *)mmmbed_name_%d, mmmbed_data_%d, %d}", i, i, len(a.marker)), fmt.Sprintf("{(const char *)mmmbed_name_%d, mmmbed_data_%d, %d}", i, i, a.size), 1))
			template = []byte(strings.Replace(string(template), fmt.Sprintf("%s_size = %d;", name[i], len(a.marker)), fmt.Sprintf("%s_size = %d;", name[i], a.size), 1))
		}
	}
	for i := range files {
		marker := files[i].marker
		switch lang {
		case "c", "cpp", "rust":
			files[i].token = numbers(marker) + "," // trailing comma belongs to the data
		case "r":
			files[i].token = numbers(marker)
		case "lua":
			s := luaString(marker)
			files[i].token = s[1 : len(s)-1]
		default:
			files[i].token = base64.StdEncoding.EncodeToString(marker)
			if lang == "java" && files[i].size > 12000 {
				files[i].token = `"` + files[i].token + `"`
			}
		}
		if strings.Count(string(template), files[i].token) != 1 {
			return fmt.Errorf("internal error: data marker for %q is ambiguous", inputs[i].Name)
		}
	}
	out := bufio.NewWriterSize(w, 64*1024)
	remaining := string(template)
	for len(remaining) > 0 {
		pos, which := -1, -1
		for i, a := range files {
			if p := strings.Index(remaining, a.token); p >= 0 && (pos < 0 || p < pos) {
				pos, which = p, i
			}
		}
		if which < 0 {
			if _, err = io.WriteString(out, remaining); err != nil {
				return err
			}
			break
		}
		if _, err = io.WriteString(out, remaining[:pos]); err != nil {
			return err
		}
		f := files[which].file
		if _, err = f.Seek(0, io.SeekStart); err != nil {
			return err
		}
		switch lang {
		case "c", "cpp", "rust", "r", "lua":
			err = writeLiteral(out, f, lang)
		case "java":
			if files[which].size > 12000 {
				err = writeJavaChunks(out, f)
			} else {
				enc := base64.NewEncoder(base64.StdEncoding, out)
				_, err = io.Copy(enc, f)
				if e := enc.Close(); err == nil {
					err = e
				}
			}
		default:
			enc := base64.NewEncoder(base64.StdEncoding, out)
			_, err = io.Copy(enc, f)
			if closeErr := enc.Close(); err == nil {
				err = closeErr
			}
		}
		if err != nil {
			return err
		}
		remaining = remaining[pos+len(files[which].token):]
	}
	return out.Flush()
}

func writeLiteral(w io.Writer, r io.Reader, lang string) error {
	input := make([]byte, 32*1024)
	output := make([]byte, 0, 160*1024)
	first := true
	for {
		n, err := r.Read(input)
		output = output[:0]
		for _, v := range input[:n] {
			switch lang {
			case "lua":
				output = append(output, '\\', '0'+v/100, '0'+(v/10)%10, '0'+v%10)
			case "r":
				if !first {
					output = append(output, ',')
				}
				output = strconv.AppendInt(output, int64(v), 10)
			default:
				output = strconv.AppendInt(output, int64(v), 10)
				output = append(output, ',')
			}
			first = false
		}
		if len(output) > 0 {
			if _, e := w.Write(output); e != nil {
				return e
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrNoProgress
		}
	}
}

// Java limits individual constant-pool strings to 65535 bytes. Use smaller
// non-folded chunks and join at runtime to stay below that limit.
func writeJavaChunks(w io.Writer, r io.Reader) error {
	if _, err := io.WriteString(w, "String.join(\"\""); err != nil {
		return err
	}
	buf := make([]byte, 12000)
	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			if _, e := io.WriteString(w, ", "+strconv.Quote(base64.StdEncoding.EncodeToString(buf[:n]))); e != nil {
				return e
			}
		}
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, ")")
	return err
}
