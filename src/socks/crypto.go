package main

import (
	"crypto/cipher"
	"crypto/rc4"
	"io"
)

type CipherStream struct {
	reader      io.Reader
	writeCloser io.WriteCloser
}

func NewCipherStream(rwc io.ReadWriteCloser, cryptMethod string, password []byte) (*CipherStream, error) {
	var stream *CipherStream
	switch cryptMethod {
	default:
		stream = &CipherStream{
			reader:      rwc,
			writeCloser: rwc,
		}

	case "rc4":
		{
			rc4CipherRead, err := rc4.NewCipher(password)
			if err != nil {
				return nil, err
			}
			rc4CipherWrite, err := rc4.NewCipher(password)
			if err != nil {
				return nil, err
			}
			stream = &CipherStream{
				reader: &cipher.StreamReader{
					S: rc4CipherRead,
					R: rwc,
				},
				writeCloser: &cipher.StreamWriter{
					S: rc4CipherWrite,
					W: rwc,
				},
			}
		}
	}
	return stream, nil
}

func (c *CipherStream) Read(data []byte) (int, error) {
	return c.reader.Read(data)
}

func (c *CipherStream) Write(data []byte) (int, error) {
	return c.writeCloser.Write(data)
}

func (c *CipherStream) Close() error {
	return c.writeCloser.Close()
}
