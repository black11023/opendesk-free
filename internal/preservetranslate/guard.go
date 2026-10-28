// Package preservetranslate implements fail-closed protection of literal
// numbers and identifiers before translation. It does NOT translate files.
// Keep the audit / form-preservation pipeline separate from language engines.
package preservetranslate

import (
    "errors"
    "fmt"
    "regexp"
    "strings"
)

const maxProtectedSpans = 4096
const markerPrefix = "__ODS_LOCK_"

// candidate matches number-bearing items, dates, currency amounts, percentages,
// and identifiers such as DN20 or PPR-DN20. More complex business keys are
// protected by a future explicit glossary / field whitelist, not inferred here.
var candidate = regexp.MustCompile(`(?:[$€£¥₫])?[\p{L}_/-]*\d[\p{L}\d.,%/:_-]*`)
var markerRE = regexp.MustCompile(`__ODS_LOCK_[0-9]{4}__`)

type LockedSpan struct {
    Marker string
    Original string
    StartByte int
    EndByte int
}

// Protect replaces numeric/code spans with unique markers; it never edits
// source bytes or claims to understand the surrounding language.
func Protect(source string) (string, []LockedSpan, error) {
    if strings.Contains(source, markerPrefix) {
        return "", nil, errors.New("input already contains OpenDesk lock markers")
    }
    indices := candidate.FindAllStringIndex(source, maxProtectedSpans+1)
    if len(indices)>maxProtectedSpans {
        return "", nil, fmt.Errorf("too many protected spans: maximum %d", maxProtectedSpans)
    }
    var b strings.Builder
    b.Grow(len(source)+len(indices)*18)
    locked := make([]LockedSpan,0,len(indices))
    prev:=0
    for n, pair := range indices {
        if pair[0]<prev || pair[1]>len(source) {return "",nil,errors.New("invalid protected span")}
        marker:=fmt.Sprintf("%s%04d__",markerPrefix,n+1)
        b.WriteString(source[prev:pair[0]])
        b.WriteString(marker)
        locked=append(locked,LockedSpan{Marker:marker,Original:source[pair[0]:pair[1]],StartByte:pair[0],EndByte:pair[1]})
        prev=pair[1]
    }
    b.WriteString(source[prev:])
    return b.String(),locked,nil
}

// Restore rejects missing, duplicated, reordered, or injected placeholders.
// That failure must be sent for human review; silently filling missing tokens
// or guessing quantities/codes is forbidden.
func Restore(translated string, spans []LockedSpan) (string,error) {
    tokens:=markerRE.FindAllString(translated,-1)
    if len(tokens)!=len(spans) {
        return "",fmt.Errorf("protected token count mismatch: got %d, want %d",len(tokens),len(spans))
    }
    if len(spans)==0 {
        if strings.Contains(translated,markerPrefix) {return "",errors.New("unexpected lock marker")}
        return translated,nil
    }
    var ordered []string
    for _,s:=range spans {
        if s.Marker==""||!markerRE.MatchString(s.Marker) || strings.Contains(s.Original,markerPrefix) {
            return "",errors.New("invalid protected span")
        }
        ordered=append(ordered,s.Marker)
    }
    for i,t:=range tokens {
        if t!=ordered[i] {
            return "",fmt.Errorf("protected token changed or reordered at position %d",i+1)
        }
    }
    out:=translated
    for _,s:=range spans {out=strings.Replace(out,s.Marker,s.Original,1)}
    if strings.Contains(out,markerPrefix) {return "",errors.New("stray lock marker in translated output")}
    return out,nil
}
