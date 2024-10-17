package utils

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

func ConvertToStringSlice[T ~string](status []T) []string {
	strSlice := make([]string, len(status))
	for i, s := range status {
		strSlice[i] = string(s)
	}
	return strSlice
}

const (
	ptrSize    = unsafe.Sizeof(uintptr(0))
	sliceSize  = unsafe.Sizeof([]int{})
	stringSize = unsafe.Sizeof("")
)

type FieldSize struct {
	Name   string
	Size   uintptr
	Count  int
	Fields []FieldSize
}

type MemoryContext struct {
	SeenPtrs map[uintptr]bool
}

func NewMemoryContext() *MemoryContext {
	return &MemoryContext{
		SeenPtrs: make(map[uintptr]bool),
	}
}

func analyzeMemoryUsage(s interface{}) FieldSize {
	v := reflect.ValueOf(s)
	ctx := NewMemoryContext()
	return calculateSizeWithBreakdown(v, "root", ctx)
}

func calculateSizeWithBreakdown(v reflect.Value, name string, ctx *MemoryContext) FieldSize {
	switch v.Kind() {
	case reflect.Struct:
		return calculateStructSizeWithBreakdown(v, name, ctx)
	case reflect.Array, reflect.Slice:
		return calculateArrayOrSliceSizeWithBreakdown(v, name, ctx)
	case reflect.Map:
		return calculateMapSizeWithBreakdown(v, name, ctx)
	case reflect.Ptr:
		return calculatePointerSize(v, name, ctx)
	default:
		return FieldSize{Name: name, Size: calculateSize(v, ctx), Count: 1}
	}
}

func calculateStructSizeWithBreakdown(v reflect.Value, name string, ctx *MemoryContext) FieldSize {
	fieldSize := FieldSize{Name: name, Count: 1}
	for i := 0; i < v.NumField(); i++ {
		subField := v.Field(i)
		subFieldName := v.Type().Field(i).Name
		subFieldSize := calculateSizeWithBreakdown(subField, subFieldName, ctx)
		fieldSize.Size += subFieldSize.Size
		fieldSize.Fields = append(fieldSize.Fields, subFieldSize)
	}
	return fieldSize
}

func calculateArrayOrSliceSizeWithBreakdown(v reflect.Value, name string, ctx *MemoryContext) FieldSize {
	fieldSize := FieldSize{Name: name, Count: v.Len()}
	if v.Len() == 0 {
		return fieldSize
	}

	elemType := v.Type().Elem()
	if elemType.Kind() == reflect.Struct {
		cumulativeSizes := make(map[string]*FieldSize)
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			for j := 0; j < elem.NumField(); j++ {
				field := elem.Field(j)
				fieldName := elemType.Field(j).Name
				subFieldSize := calculateSizeWithBreakdown(field, fieldName, ctx)

				if existing, ok := cumulativeSizes[fieldName]; ok {
					existing.Size += subFieldSize.Size
					existing.Count += subFieldSize.Count
					// Merge subfields if they exist
					if len(subFieldSize.Fields) > 0 {
						existing.Fields = mergeFields(existing.Fields, subFieldSize.Fields)
					}
				} else {
					cumulativeSizes[fieldName] = &subFieldSize
				}

				fieldSize.Size += subFieldSize.Size
			}
		}

		for _, cs := range cumulativeSizes {
			cs.Name = fmt.Sprintf("%s (%d items)", cs.Name, cs.Count)
			fieldSize.Fields = append(fieldSize.Fields, *cs)
		}
	} else {
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			fieldSize.Size += calculateSize(elem, ctx)
		}
	}

	if v.Kind() == reflect.Slice {
		fieldSize.Size += sliceSize
	}

	return fieldSize
}

func mergeFields(existing, new []FieldSize) []FieldSize {
	if len(existing) == 0 {
		return new
	}
	for i, field := range new {
		if i < len(existing) {
			existing[i].Size += field.Size
			existing[i].Count += field.Count
			existing[i].Fields = mergeFields(existing[i].Fields, field.Fields)
		} else {
			existing = append(existing, field)
		}
	}
	return existing
}

func calculateMapSizeWithBreakdown(v reflect.Value, name string, ctx *MemoryContext) FieldSize {
	fieldSize := FieldSize{Name: name, Count: v.Len()}
	if v.Len() == 0 {
		return fieldSize
	}

	valueType := v.Type().Elem()
	if valueType.Kind() == reflect.Struct {
		cumulativeSizes := make(map[string]*FieldSize)
		for _, key := range v.MapKeys() {
			elem := v.MapIndex(key)
			fieldSize.Size += calculateSize(key, ctx)
			for i := 0; i < elem.NumField(); i++ {
				field := elem.Field(i)
				fieldName := valueType.Field(i).Name
				subFieldSize := calculateSizeWithBreakdown(field, fieldName, ctx)

				if existing, ok := cumulativeSizes[fieldName]; ok {
					existing.Size += subFieldSize.Size
					existing.Count += subFieldSize.Count
					existing.Fields = mergeFields(existing.Fields, subFieldSize.Fields)
				} else {
					cumulativeSizes[fieldName] = &subFieldSize
				}

				fieldSize.Size += subFieldSize.Size
			}
		}

		for _, cs := range cumulativeSizes {
			cs.Name = fmt.Sprintf("%s (%d items)", cs.Name, cs.Count)
			fieldSize.Fields = append(fieldSize.Fields, *cs)
		}
	} else {
		for _, key := range v.MapKeys() {
			fieldSize.Size += calculateSize(key, ctx) + calculateSize(v.MapIndex(key), ctx)
		}
	}

	fieldSize.Size += uintptr(v.Len()) * ptrSize // Approximate overhead for map internals

	return fieldSize
}

func calculatePointerSize(v reflect.Value, name string, ctx *MemoryContext) FieldSize {
	fieldSize := FieldSize{Name: name, Count: 1}
	if v.IsNil() {
		fieldSize.Size = ptrSize
		return fieldSize
	}

	addr := v.Pointer()
	if ctx.SeenPtrs[addr] {
		fieldSize.Size = ptrSize
		return fieldSize
	}

	ctx.SeenPtrs[addr] = true
	elemSize := calculateSizeWithBreakdown(v.Elem(), name, ctx)
	fieldSize.Size = ptrSize + elemSize.Size
	fieldSize.Fields = elemSize.Fields
	return fieldSize
}

func calculateSize(v reflect.Value, ctx *MemoryContext) uintptr {
	switch v.Kind() {
	case reflect.Array:
		return calculateArraySize(v, ctx)
	case reflect.Slice:
		return calculateSliceSize(v, ctx)
	case reflect.String:
		return stringSize + uintptr(v.Len())
	case reflect.Map:
		return calculateMapSize(v, ctx)
	case reflect.Ptr:
		return calculatePointerSizeSimple(v, ctx)
	case reflect.Interface:
		if v.IsNil() {
			return ptrSize
		}
		return ptrSize + calculateSize(v.Elem(), ctx)
	case reflect.Struct:
		return calculateStructSize(v, ctx)
	default:
		return v.Type().Size()
	}
}

func calculateArraySize(v reflect.Value, ctx *MemoryContext) uintptr {
	totalSize := uintptr(0)
	for i := 0; i < v.Len(); i++ {
		totalSize += calculateSize(v.Index(i), ctx)
	}
	return totalSize
}

func calculateSliceSize(v reflect.Value, ctx *MemoryContext) uintptr {
	if v.IsNil() {
		return 0
	}
	return sliceSize + calculateArraySize(v, ctx)
}

func calculateMapSize(v reflect.Value, ctx *MemoryContext) uintptr {
	if v.IsNil() {
		return 0
	}
	totalSize := uintptr(0)
	for _, key := range v.MapKeys() {
		totalSize += calculateSize(key, ctx) + calculateSize(v.MapIndex(key), ctx)
	}
	return totalSize + uintptr(v.Len())*ptrSize // Approximate overhead for map internals
}

func calculateStructSize(v reflect.Value, ctx *MemoryContext) uintptr {
	totalSize := uintptr(0)
	for i := 0; i < v.NumField(); i++ {
		totalSize += calculateSize(v.Field(i), ctx)
	}
	return totalSize
}

func calculatePointerSizeSimple(v reflect.Value, ctx *MemoryContext) uintptr {
	if v.IsNil() {
		return ptrSize
	}

	addr := v.Pointer()
	if ctx.SeenPtrs[addr] {
		return ptrSize
	}

	ctx.SeenPtrs[addr] = true
	return ptrSize + calculateSize(v.Elem(), ctx)
}

// HumanizeBytes converts a byte size to a human-readable string
func humanizeBytes(bytes uintptr) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func PrintFieldSizes(s interface{}) string {
	allFields := analyzeMemoryUsage(s)
	msg := printFieldSizesRecursive([]FieldSize{allFields}, 0)
	fmt.Print(msg)
	return msg
}

func printFieldSizesRecursive(fields []FieldSize, indent int) string {
	msg := ""
	for _, field := range fields {
		msg += fmt.Sprintf("%s%s: %s\n", strings.Repeat("  ", indent), field.Name, humanizeBytes(field.Size))
		if len(field.Fields) > 0 {
			msg += printFieldSizesRecursive(field.Fields, indent+1)
		}
	}
	return msg
}
