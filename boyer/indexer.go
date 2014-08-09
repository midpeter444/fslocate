package boyer

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var verbose bool

const (
	OUT_FILE    = "db/fslocate.boyer"
	INDEX_FILE  = "conf/fslocate.indexlist"
	IGNORE_FILE = "conf/fslocate.ignore"
	PATH_SEP    = string(os.PathSeparator)
	BUFSZ       = 2097152 // 2MiB cache before flush to disk
	RECORD_SEP  = 0x1e    // "Record Separator" char in ASCII
)

type BoyerFsLocate struct{}

type ignorePatterns struct {
	suffixes []string
	patterns []string
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

/* ---[ INDEX ]--- */

func (_ BoyerFsLocate) Index(numIndexes int, beVerbose bool) {
	verbose = beVerbose

	tmpOut := OUT_FILE + randVal()
	prn("Temp out file: " + tmpOut)
	file, err := os.Create(tmpOut)
	if err != nil {
		log.Fatalf("ERROR: %v\n", err)
	}
	defer os.Remove(tmpOut)
	defer file.Close()

	dirChan := make(chan string, 10000)
	getToplevelEntries(dirChan)
	prf("Read in %d top level entries\n", len(dirChan))
	ignorePats := readInIgnorePatterns()

	var buf bytes.Buffer
	for len(dirChan) > 0 {
		dir := <-dirChan
		prf("Procesing dir: %s\n", dir)
		err := writeEntry(&buf, file, dir)
		if err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}

		var entries []os.FileInfo
		entries, err = ioutil.ReadDir(dir)
		if err != nil {
			log.Fatalf("ERROR: %v\n", err)
		}

		// processEntries(dir, entries, ignorePats, dirChan, file)
		for _, e := range entries {
			fullpath := createFullPath(dir, e.Name())
			if !shouldIgnore(ignorePats, fullpath) {
				if e.IsDir() {
					dirChan <- fullpath
				} else {
					prf("Writing entry: %s\n", fullpath)
					err := writeEntry(&buf, file, fullpath)
					if err != nil {
						log.Fatalf("ERROR: %v\n", err)
					}
				}
			}
		}
	}

	prf("buf len before final pad to limit: %d\n", buf.Len())
	padToLimit(&buf)
	prf("buf len *after final pad to limit: %d\n", buf.Len())
	flushBuffer(&buf, file)
	file.Close()
	os.Remove(OUT_FILE)
	err = os.Rename(tmpOut, OUT_FILE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Unable to copy new boyer db to %s: %v\n", OUT_FILE, err)
		return
	}
}

// TODO: haven't dealt with case where len(entry) > BUFSZ
func writeEntry(buf *bytes.Buffer, file *os.File, entry string) error {
	// +1 to add in the size of the record separator char
	if buf.Len()+len(entry)+1 > BUFSZ {
		prf("writeEntry: PadToLimit called for entry: %s\n", entry)
		padToLimit(buf)
		flushBuffer(buf, file)
	}

	_, err := buf.WriteString(entry)
	if err != nil {
		return err
	}
	_, err = buf.WriteRune(RECORD_SEP)
	if err != nil {
		return err
	}

	if buf.Len() == BUFSZ {
		flushBuffer(buf, file)
	}
	return nil
}

func padToLimit(buf *bytes.Buffer) {
	var diff = BUFSZ - buf.Len()
	for i := 0; i < diff; i++ {
		buf.WriteRune(RECORD_SEP)
	}
}

func flushBuffer(buf *bytes.Buffer, file *os.File) error {
	_, err := file.WriteString(buf.String()) // TODO: write buf.Bytes instead?  more performant?
	file.Sync()
	buf.Reset()
	return err
}

//
// Reads in the ingore patterns from IGNORE_FILE
// and returns the entries as an ignorePatterns struct
//
func readInIgnorePatterns() *ignorePatterns {
	var suffixes, patterns []string

	if !fileExists(IGNORE_FILE) {
		fmt.Fprintf(os.Stderr, "WARN: Unable to find ignore patterns file: %v\n", IGNORE_FILE)
		return nil
	}

	file, err := os.Open(IGNORE_FILE)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARN: Unable to open file for reading: %v\n", IGNORE_FILE)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ln := strings.TrimSpace(scanner.Text())
		if len(ln) != 0 && !strings.HasPrefix(ln, "#") {
			suffixes, patterns = categorizeIgnorePattern(suffixes, patterns, ln)
		}
	}

	if err = scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "WARN: Error reading in %v: %v\n", IGNORE_FILE, err)
	}
	return &ignorePatterns{suffixes: suffixes, patterns: patterns}
}

//
// Uses the ignore patterns to determine if the file/dir passed in should
// not be indexed. The full path (abspath) is checked as a pure string match first.
// If that is not found in the ignore patterns, then a regex based search is done (??)
//
func shouldIgnore(ignore *ignorePatterns, abspath string) bool {
	if ignore == nil {
		return false
	}
	for _, suffix := range ignore.suffixes {
		if strings.HasSuffix(abspath, suffix) {
			return true
		}
	}

	for _, pat := range ignore.patterns {
		if strings.Contains(abspath, pat) {
			return true
		}
	}
	return false
}

func createFullPath(dir, fname string) string {
	var buf bytes.Buffer
	buf.WriteString(dir)
	buf.WriteRune(os.PathSeparator)
	buf.WriteString(fname)
	return buf.String()
}

func getToplevelEntries(ch chan string) {
	if !fileExists(INDEX_FILE) {
		log.Fatal("ERROR: Cannot find file " + INDEX_FILE)
	}

	file, err := os.Open(INDEX_FILE)
	if err != nil {
		log.Fatal("ERROR: Cannot open file " + INDEX_FILE)
	}
	defer file.Close()

	scnr := bufio.NewScanner(file)
	for scnr.Scan() {
		ln := strings.TrimSpace(scnr.Text())
		if len(ln) != 0 && !strings.HasPrefix(ln, "#") {
			ch <- ln
		}
	}
	if err = scnr.Err(); err != nil {
		log.Fatalf("ERROR while reading %s: %v\n", INDEX_FILE, err)
	}
}

func categorizeIgnorePattern(suffixes, patterns []string, token string) ([]string, []string) {
	tok := token
	if strings.HasPrefix(tok, "*") {
		tok = tok[1:]
		suffixes = append(suffixes, tok)
	} else if strings.HasSuffix(tok, "/") {
		suffixes = append(suffixes, ensurePrefix(tok[:len(tok)-1], "/"))
		patterns = append(patterns, ensurePrefix(tok, "/"))
	} else {
		patterns = append(patterns, ensurePrefix(tok, "/"))
	}
	return suffixes, patterns
}

func ensurePrefix(s string, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s
	}
	return prefix + s
}

func randVal() string {
	n := rand.Intn(9999999999)
	return strconv.Itoa(n)
}

func fileExists(fpath string) bool {
	_, err := os.Stat(fpath)
	return err == nil
}

/* ---[ helpers ]--- */

func pr(s string) {
	if verbose {
		fmt.Print(s)
		os.Stdout.Sync()
	}
}

func prn(s string) {
	if verbose {
		fmt.Println(s)
		os.Stdout.Sync()
	}
}

func prf(format string, vals ...interface{}) {
	if verbose {
		fmt.Printf(format, vals...)
		os.Stdout.Sync()
	}
}
