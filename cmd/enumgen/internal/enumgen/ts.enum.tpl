import { Dict } from './dictDef';

export enum DictType {
{{- range $e := .Enums}}
    {{$e.TypeName}} = '{{styleName $.TypeStyle $e.TypeName}}', {{if $e.Explain}}// {{$e.Explain}}{{- end}}
{{- end}}
}

export const ENUMS = {
    {{- range $e := .Enums}}
    {{$e.TypeName}}: {
    {{- range $ee := .Values}}
        {{ formatTsEnumValue $ee.OriginalName $e.TypeName}}: {{if $ee.IsString}}'{{$ee.RawValue}}'{{else}}{{$ee.RawValue}}{{end}}, // {{$ee.Label}}
    {{- end}}
    } as const,
    {{- end}}
} as const;

export const Enums = {
{{- range $e := .Enums}}
{{- if $e.TypeComment}}
    // {{$e.TypeComment}}
{{- end}}
    {{$e.TypeName}}: new Dict('{{styleName $.TypeStyle $e.TypeName}}', [
    {{- range $ee := .Values}}
        { value: {{ printf "ENUMS.%s.%s" $e.TypeName (formatTsEnumValue $ee.OriginalName $e.TypeName) }}, label: '{{$ee.Label}}' }, // {{$ee.RawValue}}: {{$ee.Label}}
    {{- end}}
    ]),
{{- end}}
}
