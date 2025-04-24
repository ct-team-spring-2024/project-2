package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	line := strings.TrimSpace(scanner.Text())
	parts := strings.Split(line, " ")

	if len(parts) != 3 {
		fmt.Println("Error: Please provide exactly 3 numbers separated by spaces")
		return
	}

	fmt.Printf("Input values: %v\n", parts)

	num1, _ := strconv.Atoi(parts[0])
	num2, _ := strconv.Atoi(parts[1])
	megabytes, _ := strconv.Atoi(parts[2])

	memory := make([]byte, megabytes*1024*1024)
	for i := range memory {
		memory[i] = byte(i % 8)
	}

	sum := num1 / num2

	time.Sleep(2 * time.Second)
	fmt.Printf("Sum: %d\n", sum)
}
