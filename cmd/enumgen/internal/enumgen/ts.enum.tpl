import { Dict } from './dict';

enum DictType {
{{- range $e := .Enums}}
    {{$e.TypeName}} = '{{styleName $.TypeStyle $e.TypeName}}', {{if $e.Explain}}// {{$e.Explain}}{{- end}}
{{- end}}
}

export default {
    DictType,
{{- range $e := .Enums}}
{{- if $e.Explain}}
    // {{$e.Explain}}
{{- end}}
    {{$e.TypeName}}: new Dict('{{styleName $.TypeStyle $e.TypeName}}', [
    {{- range $ee := .Values}}
        { value: {{if $ee.IsString}}'{{$ee.RawValue}}'{{else}}{{$ee.RawValue}}{{end}}, label: '{{$ee.Label}}' },
    {{- end}}
    ]),
{{- end}}
}
