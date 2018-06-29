package logyzer

import (
    "bufio"
    "fmt"
    "io"
    "time"
    "os"
)

// Tail is simulating the 'tail -f' command. Needs three parameters:
// 1. The name of the file to be read.
// 2. The channel where all of the lines will be sent.
// 3. The channel where send when it's finished.
func Tail(name string, channel chan string, done chan bool) error {
    file, err := os.Open(name)
    if err != nil {
        done <- true
        return fmt.Errorf("Unable to open file %s: %s", name, err)
    }
    defer file.Close()
    r := bufio.NewReader(file)
    for {
        line, err := r.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                time.Sleep(time.Second)
            } else {
                done <- true
                return err
            }
        } else {
            channel <- line
        }
    }
    return nil
}
