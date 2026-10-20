package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

// The model of the declarations.
//
// The reference payloads of the Ministry and the schemas that describe them
// are read once, and what they hold becomes the model the generator writes
// the Go code from.

// fieldInfo describes one field of a declaration.
type fieldInfo struct {
	// nested tells whether the field holds an object of its own.
	nested bool

	// innerKey is the name the API gives to the inner object of a nested
	// object: it names it after the declaration type.
	innerKey string

	// inner is the inner object of a nested field.
	inner *jsonValue

	// entries are the entries of a nested list.
	entries []*jsonValue

	// value is the value of a plain field.
	value *jsonValue
}

// field is a field of a declaration and what is known of it.
type field struct {
	name string
	info *fieldInfo
}

// variant is a declaration type of a family.
type variant struct {
	family string
	code   string
	sample string
	xsd    string
	fields []field
	root   string
	inner  string
}

// group is a family of declarations: the fields every type of it carries and
// the fields only some of them carry.
type group struct {
	spec     groupSpec
	variants []*variant
	union    []string
	owner    map[string][]string
	shared   []string
	optional []string
}

// generator writes the Go code of the declarations.
type generator struct {
	// root is the directory of the module.
	root string

	// samples is the directory of the reference payloads.
	samples string

	// schemas is the directory of the schemas of the Ministry.
	schemas string

	// types holds the kind of every field, per schema.
	types schemaTypes

	// docs holds the official description of every field.
	docs fieldDocs

	// groups holds the families of declarations.
	groups []*group

	// usedOptional holds the kinds of the optional values the models use.
	usedOptional map[string]bool
}

// read reads the schemas and the reference payloads, and works out the
// model of every family of declarations.
func (g *generator) read() error {
	types, docs, err := readSchemas(g.schemas)
	if err != nil {
		return err
	}
	g.types = types
	g.docs = docs
	g.usedOptional = map[string]bool{}

	for _, spec := range groupSpecs {
		entry, err := g.readGroup(spec)
		if err != nil {
			return err
		}
		g.groups = append(g.groups, entry)
	}

	return nil
}

// readGroup reads the reference payloads of a family of declarations.
func (g *generator) readGroup(spec groupSpec) (*group, error) {
	entry := &group{spec: spec, owner: map[string][]string{}}

	for _, item := range spec.variants {
		path := filepath.Join(g.samples, item.sample)

		fields, root, inner, err := readFields(path)
		if err != nil {
			return nil, err
		}

		entry.variants = append(entry.variants, &variant{
			family: item.family,
			code:   item.code,
			sample: item.sample,
			xsd:    item.xsd,
			fields: fields,
			root:   root,
			inner:  inner,
		})
	}

	// The fields of the family are the fields of every type of it, in the
	// order they are written in.
	for _, item := range entry.variants {
		for _, field := range item.fields {
			if !containsName(entry.union, field.name) {
				entry.union = append(entry.union, field.name)
			}
			if !hasFamily(entry.owner[field.name], item.family) {
				entry.owner[field.name] = append(entry.owner[field.name], item.family)
			}
		}
	}

	// A field every type of the family carries is shared, and a field only
	// some of them carry is optional.
	for _, name := range entry.union {
		shared := true
		for _, item := range entry.variants {
			if !containsField(item.fields, name) {
				shared = false

				break
			}
		}

		if shared {
			entry.shared = append(entry.shared, name)
		} else {
			entry.optional = append(entry.optional, name)
		}
	}

	return entry, nil
}

// readFields reads the fields of a reference payload, and the keys of its
// envelope.
//
// A field whose value is an object is a nested object. The API names its
// inner object after the declaration type, so the inner key is kept with
// the field and the public type does not carry it.
func readFields(path string) ([]field, string, string, error) {
	document, err := readJSONFile(path)
	if err != nil {
		return nil, "", "", err
	}

	root := document.firstKey()
	inner := document.get(root)
	innerKey := inner.firstKey()

	value := inner
	if innerKey != "" {
		value = inner.get(innerKey)
	}

	if value.isList() {
		if len(value.list) > 0 {
			value = value.list[0]
		} else {
			value = &jsonValue{kind: jsonObject}
		}
	}
	if !value.isObject() {
		value = &jsonValue{kind: jsonObject}
	}

	fields := make([]field, 0, len(value.members))
	for _, member := range value.members {
		item := member.value

		if !item.isEmpty() {
			nestedKey := item.firstKey()
			nested := item.get(nestedKey)

			info := &fieldInfo{nested: true, innerKey: nestedKey, inner: nested}
			if nested.isList() {
				for _, element := range nested.list {
					if element.isObject() {
						info.entries = append(info.entries, element)
					}
				}
			}

			fields = append(fields, field{name: member.name, info: info})

			continue
		}

		fields = append(fields, field{name: member.name, info: &fieldInfo{value: item}})
	}

	return fields, root, innerKey, nil
}

// innerFields returns the fields of the inner object of a nested field.
//
// The entries of a list do not always carry the same fields, so the inner
// object is the union of them.
func innerFields(info *fieldInfo) []field {
	if info == nil {
		return nil
	}

	var fields []field

	if info.inner.isList() {
		for _, element := range info.inner.list {
			if !element.isObject() {
				continue
			}

			for _, member := range element.members {
				if containsField(fields, member.name) {
					continue
				}
				fields = append(fields, field{name: member.name, info: nestedInfo(member.value)})
			}
		}

		return fields
	}

	if !info.inner.isObject() {
		return nil
	}

	for _, member := range info.inner.members {
		fields = append(fields, field{name: member.name, info: nestedInfo(member.value)})
	}

	return fields
}

// nestedInfo describes the value of a field of a nested object.
func nestedInfo(value *jsonValue) *fieldInfo {
	if !value.isEmpty() {
		return &fieldInfo{
			nested:   true,
			innerKey: value.firstKey(),
			inner:    value.get(value.firstKey()),
		}
	}

	return &fieldInfo{value: value}
}

// fieldKind returns the kind of a field and the name it carries in the
// public API.
func (g *generator) fieldKind(name string, xsd string, info *fieldInfo, optional bool) (string, string) {
	var base, public string

	switch {
	case info != nil && info.nested:
		if info.inner.isList() {
			base = "nested:List"
		} else {
			base = "nested:Object"
		}
		public = goName(name)
	default:
		if known, ok := knownFields[name]; ok {
			base, public = known.kind, known.name
		} else {
			base = "string"
			if kinds, ok := g.types[xsd]; ok {
				if kind, ok := kinds[name]; ok {
					base = kind
				}
			}
			public = goName(name)
		}
	}

	// A schema that allows one to three digits does not say how many the
	// API writes, so the reference payload decides.
	if base == "count" || base == "tally" {
		if info != nil && info.value != nil {
			if text, ok := info.value.asText(); ok && isDigits(text) {
				if len(text) == 3 {
					base = "tally"
				} else {
					base = "count"
				}
			}
		}
	}

	if optional {
		return optionalPrefix + base, public
	}

	return base, public
}

// goName returns the name of a field in the public API, when it is not
// known.
func goName(field string) string {
	field = strings.TrimPrefix(field, "f_")

	var name strings.Builder
	for _, part := range strings.Split(field, "_") {
		if part == "" {
			continue
		}

		if initialism, ok := initialisms[part]; ok {
			name.WriteString(initialism)

			continue
		}

		name.WriteString(strings.ToUpper(part[:1]) + part[1:])
	}

	return name.String()
}

// nestedName returns the name of the public type of a nested object.
func nestedName(kind, field string) string {
	if name, ok := nestedNames[kind+"/"+field]; ok {
		return name
	}

	return goName(field)
}

// wireNestedName returns the name of the wire type of a nested object.
func wireNestedName(kind, field string) string {
	return kind + goName(field)
}

// nestedEntry returns the nested object of a field, if it has one.
func (g *generator) nestedEntry(entry *group, name string) *fieldInfo {
	for _, item := range entry.variants {
		for _, field := range item.fields {
			if field.name == name && field.info != nil && field.info.nested {
				return field.info
			}
		}
	}

	return nil
}

// isList tells whether a nested object is a list.
func isList(info *fieldInfo) bool {
	return info != nil && info.inner.isList()
}

// nestedVariants returns the inner fields of a nested field, per variant.
//
// The variants of a family do not always carry the same inner fields, so
// the inner object is the union of them and every variant sends the ones it
// holds.
func (g *generator) nestedVariants(entry *group, name string) map[string][]field {
	result := map[string][]field{}

	for _, item := range entry.variants {
		info := g.variantInfo(item, name)
		if info == nil || !info.nested {
			continue
		}

		if len(info.entries) > 0 {
			// The entries of a list do not always carry the same fields,
			// so the inner object is the union of them.
			var union []field
			for _, element := range info.entries {
				for _, member := range element.members {
					if containsField(union, member.name) {
						continue
					}
					union = append(union, field{name: member.name, info: &fieldInfo{value: member.value}})
				}
			}
			result[item.code] = union

			continue
		}

		result[item.code] = innerFields(info)
	}

	return result
}

// nestedUnionFields returns the union of the inner fields of a nested
// object, with their values.
func (g *generator) nestedUnionFields(entry *group, name string) []field {
	var union []field

	for _, item := range entry.variants {
		info := g.variantInfo(item, name)
		if info == nil || !info.nested {
			continue
		}

		for _, field := range innerFields(info) {
			if containsField(union, field.name) {
				continue
			}
			union = append(union, field)
		}
	}

	return union
}

// innerPresence returns the fields of a nested object that every one of
// them carries.
//
// A field that only some of the variants, or some of the entries of a list,
// carry is optional: the API leaves it out, and the SDK has to leave it out
// too.
func (g *generator) innerPresence(entry *group, name string) map[string]bool {
	var sets []map[string]bool

	for _, item := range entry.variants {
		info := g.variantInfo(item, name)
		if info == nil || !info.nested {
			continue
		}

		if len(info.entries) > 0 {
			for _, element := range info.entries {
				names := map[string]bool{}
				for _, member := range element.members {
					names[member.name] = true
				}
				sets = append(sets, names)
			}

			continue
		}

		names := map[string]bool{}
		for _, field := range innerFields(info) {
			names[field.name] = true
		}
		sets = append(sets, names)
	}

	if len(sets) == 0 {
		return map[string]bool{}
	}

	required := map[string]bool{}
	for name := range sets[0] {
		required[name] = true
	}
	for _, names := range sets[1:] {
		for name := range required {
			if !names[name] {
				delete(required, name)
			}
		}
	}

	return required
}

// variantInfo returns what is known of a field of a variant of a family.
func (g *generator) variantInfo(item *variant, name string) *fieldInfo {
	for _, field := range item.fields {
		if field.name == name {
			return field.info
		}
	}

	return nil
}

// containsField tells whether a list of fields holds a name.
func containsField(fields []field, name string) bool {
	for _, field := range fields {
		if field.name == name {
			return true
		}
	}

	return false
}

// containsName tells whether a list of names holds a name.
func containsName(names []string, name string) bool {
	for _, item := range names {
		if item == name {
			return true
		}
	}

	return false
}

// hasFamily tells whether a list of families holds a name.
func hasFamily(families []string, name string) bool {
	for _, family := range families {
		if family == name {
			return true
		}
	}

	return false
}

// isDigits tells whether a text holds nothing but digits.
func isDigits(text string) bool {
	if text == "" {
		return false
	}

	for _, character := range text {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}

// publicType returns the type of a kind of field in the public API.
func publicType(kind string) string {
	switch {
	case strings.HasPrefix(kind, optionalPrefix):
		base := strings.TrimPrefix(kind, optionalPrefix)

		// A boolean is already a pointer, so it carries the three states on
		// its own and an optional one is the same.
		if base == "flag" {
			return "*bool"
		}

		return "*" + publicType(base)
	case strings.HasPrefix(kind, enumPrefix):
		return strings.TrimPrefix(kind, enumPrefix)
	case strings.HasPrefix(kind, "nested:"):
		return "json.RawMessage"
	default:
		return publicTypes[kind]
	}
}

// wireType returns the type of a kind of field on the wire.
func wireType(kind string) string {
	switch {
	case strings.HasPrefix(kind, optionalPrefix):
		base := strings.TrimPrefix(kind, optionalPrefix)
		if base == "flag" {
			return "*wire.Bool"
		}

		return "*" + wireType(base)
	case strings.HasPrefix(kind, enumPrefix):
		return "string"
	case strings.HasPrefix(kind, "nested:"):
		return "json.RawMessage"
	default:
		return wireTypes[kind]
	}
}

// wireTypeLocal returns the type of a kind of field on the wire, as it is
// written in the package of the wire.
func wireTypeLocal(kind string) string {
	return strings.ReplaceAll(wireType(kind), "wire.", "")
}

// convert returns the way a value of the public API is written on the wire.
func (g *generator) convert(kind string) string {
	switch {
	case strings.HasPrefix(kind, optionalPrefix):
		base := strings.TrimPrefix(kind, optionalPrefix)
		g.usedOptional[base] = true

		if base == "flag" {
			return "optBoolPtr({v})"
		}
		if strings.HasPrefix(base, enumPrefix) {
			return "optString((*string)({v}))"
		}

		return optionalHelpers[base].forward + "({v})"
	case strings.HasPrefix(kind, enumPrefix):
		return "string({v})"
	default:
		return conversions[kind]
	}
}

// convertBack returns the way a value of the wire is read into the public
// API.
func (g *generator) convertBack(kind string) string {
	switch {
	case strings.HasPrefix(kind, optionalPrefix):
		base := strings.TrimPrefix(kind, optionalPrefix)
		g.usedOptional[base] = true

		if base == "flag" {
			return "boolPtrOf({v})"
		}
		if strings.HasPrefix(base, enumPrefix) {
			return "optString((*string)({v}))"
		}

		return optionalHelpers[base].backward + "({v})"
	case strings.HasPrefix(kind, enumPrefix):
		return strings.TrimPrefix(kind, enumPrefix) + "({v})"
	default:
		return conversionsBack[kind]
	}
}

// semanticName returns the name of the constant of a variant.
func semanticName(code string) string {
	if name, ok := semantic[code]; ok {
		return name
	}

	return code
}

// lowerName returns the name of a family, as it is written at the start of
// a word.
func lowerName(kind string) string {
	if kind == "" {
		return kind
	}

	return strings.ToLower(kind[:1]) + kind[1:]
}

// checkNames makes sure that the names of the public API are unique.
func checkNames() error {
	names := map[string]bool{}

	for _, spec := range groupSpecs {
		for _, item := range spec.variants {
			name := semanticName(item.code)
			if names[name] {
				return fmt.Errorf("ergani: the name %s is used twice", name)
			}
			names[name] = true
		}
	}

	for _, name := range nestedNames {
		if names[name] {
			return fmt.Errorf("ergani: the name %s is used twice", name)
		}
		names[name] = true
	}

	return nil
}
