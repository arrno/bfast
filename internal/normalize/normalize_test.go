package normalize

import "testing"

func TestParseOwnerRepo(t *testing.T) {
	slug, err := Parse("arrno/bfast")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if slug.Owner != "arrno" || slug.Repo != "bfast" {
		t.Fatalf("unexpected slug: %+v", slug)
	}
}

func TestParseHTTPS(t *testing.T) {
	slug, err := Parse("https://github.com/arrno/blazingly-fast.git")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if slug.Owner != "arrno" || slug.Repo != "blazingly-fast" {
		t.Fatalf("unexpected slug: %+v", slug)
	}
}

func TestParseSSH(t *testing.T) {
	slug, err := Parse("git@github.com:arrno/bfast.git")
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if slug.Owner != "arrno" || slug.Repo != "bfast" {
		t.Fatalf("unexpected slug: %+v", slug)
	}
}

func TestParseInvalid(t *testing.T) {
	if _, err := Parse("not-a-repo"); err == nil {
		t.Fatal("expected error for invalid slug")
	}
}
