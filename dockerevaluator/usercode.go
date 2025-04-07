package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"strconv"
	"strings"
)

func run() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	line := strings.TrimSpace(scanner.Text())
	parts := strings.Split(line, " ")
	fmt.Printf("KIR ?%v\n", parts)


	num1, _ := strconv.Atoi(parts[0])
	num2, _ := strconv.Atoi(parts[1])

	sum := num1 + num2

	time.Sleep(2 * time.Second)
	fmt.Printf("Sum: %d\n", sum)
}
