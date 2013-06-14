// package sudomake provides helpers to make all the makeable things
package sudomake

import (
    "fmt"
    "reflect"
    "strconv"
    "strings"
    "unsafe"
)

// desc is the description of a make operation on a struct type.
type desc struct {
    offset uintptr
    t      reflect.Type
    len    int
    cap    int
}

// holds a cache of types to make descriptions.
var descCache = make(map[reflect.Type][]*desc)

// Cache is used to cache the operations required to make a struct. Cache
// expects a pointer to a struct.
func Cache(v interface{}) {
    t := reflect.TypeOf(v).Elem()
    if t.Kind() != reflect.Struct {
        panic("sudomake: Cache not passed a pointer to a struct")
    }
    if _, ok := descCache[t]; ok {
        return
    }
    cache(t)
}

// get returns the description of operations for a type.
func get(t reflect.Type) []*desc {
    if d, ok := descCache[t]; ok {
        return d
    }
    return cache(t)
}

// cache builds a slice of desc from a type and adds it to descCache.
func cache(t reflect.Type) []*desc {
    fields := t.NumField()
    d := make([]*desc, 0, fields)
    for i := 0; i < fields; i++ {
        field := t.Field(i)
        switch field.Type.Kind() {
        case reflect.Chan, reflect.Map, reflect.Slice:
            l, c := parseParams(field.Tag.Get("sudomake"))
            d = append(d, &desc{
                offset: field.Offset,
                t:      field.Type,
                len:    l,
                cap:    c,
            })
        }
    }
    descCache[t] = d
    return d
}

// parsePrams gerabs the lenght and cap values from the struct field tag
func parseParams(tag string) (l int, c int) {
    if tag == "" {
        return
    }
    idx := strings.Index(tag, ",")
    if idx == -1 {
        l64, err := strconv.ParseInt(strings.TrimSpace(tag), 10, 32)
        if err != nil {
            panic(fmt.Sprintf("sudomake: %s", err))
        }
        l = int(l64)
    } else {
        l64, err := strconv.ParseInt(strings.TrimSpace(tag[:idx]), 10, 32)
        if err != nil {
            panic(fmt.Sprintf("sudomake: %s", err))
        }
        l = int(l64)
        c64, err := strconv.ParseInt(strings.TrimSpace(tag[idx+1:]), 10, 32)
        if err != nil {
            panic(fmt.Sprintf("sudomake: %s", err))
        }
        c = int(c64)
    }
    return
}

// Make accepts a pointer to some value you would like to have all the makeable
// fields maked on.
func Make(v interface{}) {
    val := reflect.ValueOf(v).Elem()
    dataPtr := val.UnsafeAddr()
    if val.Kind() != reflect.Struct {
        panic("sudomake: not passed in a pointer to a struct")
    }
    for _, d := range get(val.Type()) {
        switch d.t.Kind() {
        // a map value fits inside an interface
        case reflect.Map:
            mi := reflect.MakeMap(d.t).Interface()
            mva := reflect.ValueOf(&mi).Elem().InterfaceData()
            m := *(*map[int]int)(unsafe.Pointer(&mva[1]))
            loc := (*map[int]int)(unsafe.Pointer(dataPtr + d.offset))
            *loc = m
        // a chan value fits inside an interface
        case reflect.Chan:
            ci := reflect.MakeChan(d.t, d.len).Interface()
            cva := reflect.ValueOf(&ci).Elem().InterfaceData()
            c := *(*chan int)(unsafe.Pointer(&cva[1]))
            loc := (*chan int)(unsafe.Pointer(dataPtr + d.offset))
            *loc = c
        // a slice value does not fit so the data points to the slice already
        case reflect.Slice:
            si := reflect.MakeSlice(d.t, d.len, d.cap).Interface()
            sva := reflect.ValueOf(&si).Elem().InterfaceData()
            s := *(*[]int)(unsafe.Pointer(sva[1]))
            loc := (*[]int)(unsafe.Pointer(dataPtr + d.offset))
            *loc = s
        default:
            panic("sudomake: found invalid description")
        }
    }
}
