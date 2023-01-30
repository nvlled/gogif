# image/gif

A modified [image/gif](https://pkg.go.dev/image/gif)
package from the go standard library
(originally by [Andy](https://github.com/andybons)).
The change is
to allow streaming gif to a file or to any io.writer.

## Example usage

```go
e := gif.NewStreamEncoder(someFile, &StreamEncoderOptions{ })
for recording {
    image := takeScreenShot()
    delay := 1
    if err := e.Encode(image, delay, gif.DisposalNone); err != nil {
        handleError(err)
    }
}

if err := e.Close(); err != nil {
    handleError(err)
}

```
