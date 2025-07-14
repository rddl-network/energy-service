package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	dbDir := flag.String("db", "", "Path to the LevelDB database directory")
	outFile := flag.String("out", "output.json", "Output JSON file")
	flag.Parse()

	if *dbDir == "" {
		log.Fatal("Please provide the path to the LevelDB database directory using --db")
	}

	db, err := leveldb.OpenFile(*dbDir, nil)
	if err != nil {
		log.Fatalf("Failed to open LevelDB: %v", err)
	}
	defer db.Close()

	iter := db.NewIterator(nil, nil)
	defer iter.Release()

	entries := make(map[string]string)
	for iter.Next() {
		key := string(iter.Key())
		value := string(iter.Value())
		entries[key] = value
	}
	if err := iter.Error(); err != nil {
		log.Fatalf("Iterator error: %v", err)
	}

	file, err := os.Create(*outFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		log.Fatalf("Failed to encode JSON: %v", err)
	}

	fmt.Printf("Exported %d entries to %s\n", len(entries), *outFile)
}
