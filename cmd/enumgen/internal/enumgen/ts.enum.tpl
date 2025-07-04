import { Dict } from './dict';

export default {
{{- range $e := .Enums}}
{{- if $e.Explain}}
    // {{$e.Explain}}
{{- end}}
    {{$e.TypeName}}: new Dict('{{$e.TypeName}}', [
    {{- range $ee := .Values}}
        { value: {{if $ee.IsString}}'{{$ee.RawValue}}'{{else}}{{$ee.RawValue}}{{end}}, label: '{{$ee.Label}}' },
    {{- end}}
    ]),
{{- end}}
}
