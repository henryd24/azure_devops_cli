package scaffold

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// Object es un objeto JSON que conserva el orden de sus claves, para poder editar
// task.json, vss-extension.json y config/*.json sin reordenarlos.
type Object struct {
	keys  []string
	vals  map[string]json.RawMessage
	dirty map[string]bool // claves modificadas; las demás se escriben tal cual estaban
}

// ParseObject decodifica un objeto JSON conservando el orden de las claves.
func ParseObject(data []byte) (*Object, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("se esperaba un objeto JSON")
	}
	o := &Object{vals: map[string]json.RawMessage{}, dirty: map[string]bool{}}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key := keyTok.(string)
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		if _, dup := o.vals[key]; !dup {
			o.keys = append(o.keys, key)
		}
		o.vals[key] = raw
	}
	return o, nil
}

// ReadObject lee y decodifica un archivo JSON.
func ReadObject(path string) (*Object, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	o, err := ParseObject(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return o, nil
}

// Get decodifica el valor de key en v. Devuelve false si no existe.
func (o *Object) Get(key string, v any) (bool, error) {
	raw, ok := o.vals[key]
	if !ok {
		return false, nil
	}
	return true, json.Unmarshal(raw, v)
}

// Raw devuelve el valor sin decodificar.
func (o *Object) Raw(key string) (json.RawMessage, bool) {
	raw, ok := o.vals[key]
	return raw, ok
}

// Set asigna key (al final si es nueva).
func (o *Object) Set(key string, v any) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return err
	}
	raw := json.RawMessage(bytes.TrimSpace(buf.Bytes()))
	if _, ok := o.vals[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.vals[key] = raw
	if o.dirty == nil {
		o.dirty = map[string]bool{}
	}
	o.dirty[key] = true
	return nil
}

// Bytes serializa el objeto con sangría de 2 espacios. Los valores que no se
// modificaron con Set se escriben exactamente como estaban en el archivo original.
func (o *Object) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, k := range o.keys {
		kb, _ := json.Marshal(k)
		buf.WriteString("  ")
		buf.Write(kb)
		buf.WriteString(": ")
		if o.dirty[k] {
			var v bytes.Buffer
			if err := json.Indent(&v, o.vals[k], "  ", "  "); err != nil {
				return nil, err
			}
			buf.Write(compactScalarArrays(v.Bytes()))
		} else {
			buf.Write(o.vals[k])
		}
		if i < len(o.keys)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString("}\n")
	if !json.Valid(buf.Bytes()) {
		return nil, fmt.Errorf("el JSON resultante no es válido")
	}
	return buf.Bytes(), nil
}

// scalarArrayRe encuentra arreglos de valores simples que json.Indent separa en varias líneas.
var scalarArrayRe = regexp.MustCompile(`\[\n\s+((?:"(?:[^"\\\n]|\\.)*"|-?\d[\d.eE+-]*|true|false|null)(?:,\n\s+(?:"(?:[^"\\\n]|\\.)*"|-?\d[\d.eE+-]*|true|false|null))*)\n\s*\]`)

var arraySepRe = regexp.MustCompile(`,\n\s+`)

// compactScalarArrays deja en una línea los arreglos cortos de valores simples
// (["Build", "Release"]), como suelen escribirse a mano.
func compactScalarArrays(data []byte) []byte {
	return scalarArrayRe.ReplaceAllFunc(data, func(m []byte) []byte {
		inner := scalarArrayRe.FindSubmatch(m)[1]
		joined := arraySepRe.ReplaceAll(inner, []byte(", "))
		if len(joined) > 70 {
			return m
		}
		return append(append([]byte("["), joined...), ']')
	})
}

// Write guarda el objeto en path.
func (o *Object) Write(path string) error {
	data, err := o.Bytes()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
