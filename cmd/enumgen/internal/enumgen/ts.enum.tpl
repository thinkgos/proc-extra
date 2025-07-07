
interface DictEntry {
    readonly value: string | number; // 字典项value
    readonly label: string; // 字典项label
    attrType?: string; // 前端使用, primary, success, info, warning, danger
    [key: string]: any; // 自定义字段和值
}

export class Dict {
    private type: string
    private items: DictEntry[];
    private mapping: Map<string, DictEntry>;

    constructor(type: string, items: DictEntry[]) {
        this.type = type;
        this.items = items;
        this.mapping = new Map(
            items.map(item => ['' + item.value, item])
        );
    }
    public getType(): string { return this.type; }
    public getEntries(): ReadonlyArray<Readonly<DictEntry>> {
        return this.items || []
    }
    public getEntry(value: string | number): Readonly<DictEntry> | undefined {
        return this.mapping.get('' + value);
    }
    public getLabel(value: string | number, defaultValue: string = '未定义'): string {
        return this.mapping.get('' + value)?.label ?? defaultValue;
    }
    public getAttrType(value: string | number, defaultValue: string = 'info'): string {
        return this.mapping.get('' + value)?.attrType ?? defaultValue;
    }
    public patchFieldValue(map: Record<string | number, Partial<Omit<DictEntry, 'value' | 'label'>>>) {
        this.items.forEach(item => {
            const key = '' + item.value;
            const partial = map[key];
            const { label: _label, value: _value, ...patch } = partial;
            if (patch) {
                Object.assign(item, patch);
                _label;
                _value;
            }
        });
    }
}

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
