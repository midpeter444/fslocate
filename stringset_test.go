package main

import (
	"fmt"
	"testing"
)

func TestAdd(t *testing.T) {
	set := StringSet{}
	if len(set) != 0 {
		t.Errorf("len: %v", len(set))
	}

	str1 := "1"
	str2 := "2"
	str3 := "3"

	set.Add(str1)
	if len(set) != 1 {
		t.Errorf("len: %v", len(set))
	}

	set.Add(str2)
	if len(set) != 2 {
		t.Errorf("len: %v", len(set))
	}

	set.Add(str3)
	if len(set) != 3 {
		t.Errorf("len: %v", len(set))
	}

	set.Add(str2)
	if len(set) != 3 {
		t.Errorf("len: %v", len(set))
	}

	set.Add(str1)
	if len(set) != 3 {
		t.Errorf("len: %v", len(set))
	}

	set.Add(str3)
	if len(set) != 3 {
		t.Errorf("len: %v", len(set))
	}

	if _, present := set[str1]; !present {
		t.Errorf("str1 not present")
	}
	if _, present := set[str2]; !present {
		t.Errorf("str2 not present")
	}
	if _, present := set[str3]; !present {
		t.Errorf("str3 not present")
	}
}

func TestStringSetContains(t *testing.T) {
	set := StringSet{}
	if len(set) != 0 {
		t.Errorf("len: %v", len(set))
	}

	str1 := "1"
	str2 := "2"
	str3 := "3"

	set.Add(str1)
	set.Add(str2)

	if !set.Contains(str1) {
		t.Errorf("str1 not present")
	}
	if !set.Contains(str2) {
		t.Errorf("str2 not present")
	}
	if set.Contains(str3) {
		t.Errorf("str3 present")
	}
}

func TestStringSetAddAll(t *testing.T) {

	set1 := StringSet{}
	set2 := StringSet{}

	str1 := "1"
	str2 := "2"
	str3 := "3"
	str4 := "4"
	str5 := "5"
	str6 := "6"

	set1.Add(str1)
	set1.Add(str2)
	set1.Add(str3)

	set2.Add(str4)
	set2.Add(str5)
	set2.Add(str6)

	if len(set1) != 3 {
		t.Errorf("len: %v", len(set1))
	}
	if len(set2) != 3 {
		t.Errorf("len: %v", len(set2))
	}

	set3 := set1.AddAll(set2)

	if len(set1) != 6 {
		t.Errorf("len: %v", len(set1))
	}
	if len(set2) != 3 {
		t.Errorf("len: %v", len(set2))
	}
	if len(set3) != 6 {
		t.Errorf("len: %v", len(set1))
	}

	if !set1.Contains(str5) {
		t.Errorf("%+v", set1)
	}
	if set2.Contains(str2) {
		t.Errorf("%v", set2)
	}

	// set3 should just be an alias for set1
	if len(set1) != len(set3) {
		t.Errorf("set1: %v; set3: %v", set1, set3)
	}
	if fmt.Sprintf("%v\n", set1) != fmt.Sprintf("%v\n", set1) {
		t.Errorf("set1: %v; set3: %v", set1, set3)
	}
}

func TestStringSetRemove(t *testing.T) {

	set := StringSet{}

	str1 := "1"
	str2 := "2"
	str3 := "3"

	set.Add(str1)
	set.Add(str2)
	set.Add(str3)

	if !set.Contains(str1) {
		t.Errorf("str1 not present")
	}
	if !set.Contains(str2) {
		t.Errorf("str2 not present")
	}
	if !set.Contains(str3) {
		t.Errorf("str3 not present")
	}

	set.Remove(str2)

	if len(set) != 2 {
		t.Errorf("len: %v", len(set))
	}
	if !set.Contains(str1) {
		t.Errorf("str1 not present")
	}
	if set.Contains(str2) {
		t.Errorf("str2 present")
	}
	if !set.Contains(str3) {
		t.Errorf("str3 not present")
	}

	set.Remove(str1)
	set.Remove(str3)
	if len(set) != 0 {
		t.Errorf("len: %v", len(set))
	}

}

func TestIsSubset(t *testing.T) {

	set1 := StringSet{}
	set2 := StringSet{}
	set3 := StringSet{}
	set4 := StringSet{}

	str1 := "1"
	str2 := "2"
	str3 := "3"
	str4 := "4"
	str5 := "5"
	str6 := "6"

	set1.Add(str1)
	set1.Add(str2)
	set1.Add(str3)

	set2.Add(str4)
	set2.Add(str5)
	set2.Add(str6)

	set3.Add(str1)
	set3.Add(str3)

	if set1.IsSubset(set2) {
		t.Errorf("set1: %v; set2: %v", set1, set2)
	}
	if set2.IsSubset(set1) {
		t.Errorf("set1: %v; set2: %v", set1, set2)
	}

	if !set3.IsSubset(set1) {
		t.Errorf("set1: %v; set3: %v", set1, set3)
	}
	if set1.IsSubset(set3) {
		t.Errorf("set1: %v; set3: %v", set1, set3)
	}

	if !set4.IsSubset(set1) {
		t.Errorf("set1: %v; set4: %v", set1, set4)
	}
	if set1.IsSubset(set4) {
		t.Errorf("set1: %v; set4: %v", set1, set4)
	}

	fmt.Printf("%v: %s\n", set1, set2)
}
