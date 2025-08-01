package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	inputPaths := os.Args[2:]

	if command != "repack" && command != "unpack" {
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	var errors []error
	successCount := 0

	for _, inputPath := range inputPaths {
		outputPath := generateOutputPath(command, inputPath)

		err := processReplay(inputPath, outputPath, command == "repack")
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", inputPath, err)
			errors = append(errors, err)
		} else {
			fmt.Printf("Successfully %sed %s -> %s\n", command, inputPath, outputPath)
			successCount++
		}
	}

	fmt.Printf("\nProcessed %d files successfully", successCount)
	if len(errors) > 0 {
		fmt.Printf(", %d files failed", len(errors))
		os.Exit(1)
	}
	fmt.Println()
}

func printUsage() {
	binary := filepath.Base(os.Args[0])
	fmt.Println("Usage:")
	fmt.Println(binary + " repack INPUT1 [INPUT2 ...]")
	fmt.Println(binary + " unpack INPUT1 [INPUT2 ...]")
}

func generateOutputPath(command, inputPath string) string {
	switch command {
	case "repack":
		return inputPath + ".repacked"
	case "unpack":
		return inputPath + ".unpacked"
	default:
		return inputPath + ".processed"
	}
}

func processReplay(inputPath, outputPath string, compress bool) error {
	rawData, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	smBytes := 27
	sourceMapOff := int(binary.LittleEndian.Uint32(rawData[smBytes : smBytes+4]))

	svBytes := 263
	if string(rawData[svBytes:svBytes+2]) != "sv" {
		return fmt.Errorf("sv bytes marker not match")
	}

	commandOff := int(binary.LittleEndian.Uint32(rawData[svBytes+2 : svBytes+6]))
	unknownOff := int(binary.LittleEndian.Uint32(rawData[svBytes+6 : svBytes+10]))
	if commandOff-unknownOff != 4 {
		return err
	}

	sourceMapSize := sourceMapOff - commandOff

	var payload []byte
	if !compress {
		payload, err = DecompressL33t(rawData)
	} else {
		payload, err = CompressL33t(rawData[svBytes+10 : commandOff])
	}
	if err != nil {
		return err
	}

	repacked := rawData[:smBytes]
	buffer := make([]byte, 4)

	binary.LittleEndian.PutUint32(buffer, uint32(svBytes+2+8+len(payload)+sourceMapSize))
	repacked = append(repacked, buffer...)

	repacked = append(repacked, rawData[smBytes+4:svBytes+2]...)

	binary.LittleEndian.PutUint32(buffer, uint32(svBytes+2+8+len(payload)))
	repacked = append(repacked, buffer...)

	binary.LittleEndian.PutUint32(buffer, uint32(svBytes+2+4+len(payload)))
	repacked = append(repacked, buffer...)

	repacked = append(repacked, payload...)
	repacked = append(repacked, rawData[commandOff:]...)

	if err := os.WriteFile(outputPath, repacked, 0o644); err != nil {
		return err
	}

	return nil
}

func DecompressL33t(compressed []byte) ([]byte, error) {
	offset := bytes.Index(compressed, []byte{0x6c, 0x33, 0x33, 0x74})
	if offset == -1 {
		return nil, fmt.Errorf("cannot find l33t header")
	}
	size := binary.LittleEndian.Uint32(compressed[offset+4 : offset+8])

	reader, err := zlib.NewReader(bytes.NewReader(compressed[offset+8:]))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if len(decompressed) != int(size) {
		return nil, fmt.Errorf("invalid decompressed size: expected %d, got %d", size, len(decompressed))
	}

	return decompressed, nil
}

func CompressL33t(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := zlib.NewWriter(&buffer)
	if _, err := writer.Write(data); err != nil {
		writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	compressedData := buffer.Bytes()
	result := make([]byte, 0, 8+len(compressedData))
	result = append(result, 0x6c, 0x33, 0x33, 0x74)

	sizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizeBytes, uint32(len(data)))

	result = append(result, sizeBytes...)
	result = append(result, compressedData...)

	return result, nil
}
