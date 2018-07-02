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
func Tail(name string, channel chan<- string, done chan<- bool, quit <-chan bool) error {
    file, err := os.Open(name)
    if err != nil {
        done <- true
        return fmt.Errorf("Unable to open file %s: %s", name, err)
    }
    file.Seek(0, 2) // go to the EOF to read only new lines
    defer file.Close()
    r := bufio.NewReader(file)
    for {
        line, err := r.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                select {
                case <-quit:
                    done <- true
                    return nil
                case <-time.After(time.Second):
                }
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
