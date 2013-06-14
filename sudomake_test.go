package sudomake

import (
    "testing"
)

type testStruct struct {
    slicef []int    `sudomake:"2,5"`
    chanf  chan int `sudomake:"7"`
    mapf   map[int]string
    welp   int
}

func TestMakes(t *testing.T) {
    x := &testStruct{}
    Make(x)
    t.Logf("%+v", x)

    if x.mapf == nil {
        t.Errorf("failed to make map")
    } else {
        x.mapf[2] = "foo"
        if x.mapf[2] != "foo" {
            t.Errorf("failed to work with map")
        }
    }
    t.Logf("%+v", x)

    if x.chanf == nil || cap(x.chanf) != 7 {
        t.Errorf("failed to make chan")
    } else {
        x.chanf <- 5
        if <-x.chanf != 5 {
            t.Errorf("chan didn't work")
        }
    }
    t.Logf("%+v", x)

    if x.slicef == nil || len(x.slicef) != 2 || cap(x.slicef) != 5 {
        t.Errorf("failed to make slice")
    } else {
        x.slicef[0] = 5
        if x.slicef[0] != 5 {
            t.Errorf("slice didn't work")
        }
    }
    t.Logf("%+v", x)

    if x.welp != 0 {
        t.Errorf("set welp for no reason")
    }
}

func BenchmarkMake(b *testing.B) {
    x := &testStruct{}
    Cache(x)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Make(x)
    }
}

func BenchmarkGoMake(b *testing.B) {
    x := &testStruct{}
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        x.chanf = make(chan int, 7)
        x.mapf = make(map[int]string)
        x.slicef = make([]int, 2, 5)
    }
}
