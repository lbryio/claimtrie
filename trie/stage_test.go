package trie

import (
	"fmt"
	"reflect"
	"testing"
)

func TestStage_Update(t *testing.T) {
	tr1 := buildTrie(New(), pairs1())

	s1 := NewStage(tr1)
	s1.Update(Key("cook"), strValue("hello"))
	s1.Update(Key("ted"), nil)

	tr1Exp := buildTrie(New(), pairs1())

	s1Exp := buildTrie(New(), pairs1())
	s1Exp.Update(Key("cook"), strValue("hello"))
	s1Exp.Update(Key("ted"), nil)

	if !reflect.DeepEqual(tr1, tr1Exp) {
		t.Errorf("Stage.Update() tr1 != tr1Exp")
		traverse(tr1.root, Key{}, dump)
		fmt.Println("")
		traverse(tr1Exp.root, Key{}, dump)
	}
	if !reflect.DeepEqual(s1.MerkleTrie, s1Exp) {
		t.Errorf("Stage.Update() s1 != s1Exp")
		traverse(s1.root, Key{}, dump)
		fmt.Println("")
		traverse(s1Exp.root, Key{}, dump)
	}
}
