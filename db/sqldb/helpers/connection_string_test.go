package helpers_test

import (
	"bytes"
	"errors"
	"io"
	"net"
	"time"

	"code.cloudfoundry.org/bbs/db/sqldb/helpers"
	"code.cloudfoundry.org/lager/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TimeoutConn - Unit Tests (No Database Required)", Label("no-db"), func() {
	var logger lager.Logger

	BeforeEach(func() {
		var buf bytes.Buffer
		logger = lager.NewLogger("test")
		logger.RegisterSink(lager.NewWriterSink(&buf, lager.DEBUG))
	})

	Describe("Read", func() {
		It("sets a read deadline before reading", func() {
			mockConn := &mockNetConn{
				readDeadlineSetTo: nil,
			}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
				Rd:   100 * time.Millisecond,
				Wd:   200 * time.Millisecond,
			}

			before := time.Now()
			_, _ = tc.Read([]byte("test"))
			after := time.Now()

			Expect(mockConn.readDeadlineSetTo).NotTo(BeNil())
			Expect(*mockConn.readDeadlineSetTo).To(BeTemporally(">", before))
			Expect(*mockConn.readDeadlineSetTo).To(BeTemporally("<=", after.Add(100*time.Millisecond)))
		})

		It("does not set a deadline if ReadTimeout is 0", func() {
			mockConn := &mockNetConn{
				readDeadlineSetTo: nil,
			}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
				Rd:   0,
				Wd:   100 * time.Millisecond,
			}

			_, _ = tc.Read([]byte("test"))

			Expect(mockConn.readDeadlineSetTo).To(BeNil())
		})

		It("returns data from the underlying connection", func() {
			mockConn := &mockNetConn{
				data: []byte("hello"),
			}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
				Rd:   100 * time.Millisecond,
				Wd:   100 * time.Millisecond,
			}

			buf := make([]byte, 10)
			n, err := tc.Read(buf)

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(5))
			Expect(buf[:n]).To(Equal([]byte("hello")))
		})

		It("returns errors from the underlying connection", func() {
			mockConn := &mockNetConn{
				readErr: errors.New("read failed"),
			}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
				Rd:   100 * time.Millisecond,
				Wd:   100 * time.Millisecond,
			}

			buf := make([]byte, 10)
			_, err := tc.Read(buf)

			Expect(err).To(MatchError("read failed"))
		})
	})

	Describe("Write", func() {
		It("sets a write deadline before writing", func() {
			mockConn := &mockNetConn{
				readDeadlineSetTo:  nil,
				writeDeadlineSetTo: nil,
			}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
				Rd:   100 * time.Millisecond,
				Wd:   200 * time.Millisecond,
			}

			before := time.Now()
			_, _ = tc.Write([]byte("test"))
			after := time.Now()

			Expect(mockConn.writeDeadlineSetTo).NotTo(BeNil())
			Expect(*mockConn.writeDeadlineSetTo).To(BeTemporally(">", before))
			Expect(*mockConn.writeDeadlineSetTo).To(BeTemporally("<=", after.Add(200*time.Millisecond)))
		})

		It("does not set a deadline if WriteTimeout is 0", func() {
			mockConn := &mockNetConn{
				writeDeadlineSetTo: nil,
			}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
				Rd:   100 * time.Millisecond,
				Wd:   0,
			}

			_, _ = tc.Write([]byte("test"))

			Expect(mockConn.writeDeadlineSetTo).To(BeNil())
		})

		It("returns the number of bytes written", func() {
			mockConn := &mockNetConn{}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
				Rd:   100 * time.Millisecond,
				Wd:   100 * time.Millisecond,
			}

			n, err := tc.Write([]byte("hello"))

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(5))
		})

		It("returns errors from the underlying connection", func() {
			mockConn := &mockNetConn{
				writeErr: errors.New("write failed"),
			}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
				Rd:   100 * time.Millisecond,
				Wd:   100 * time.Millisecond,
			}

			_, err := tc.Write([]byte("test"))

			Expect(err).To(MatchError("write failed"))
		})
	})

	Describe("other net.Conn methods", func() {
		It("delegates to underlying connection for Close", func() {
			mockConn := &mockNetConn{}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
			}

			err := tc.Close()

			Expect(mockConn.closeCalled).To(BeTrue())
			Expect(err).NotTo(HaveOccurred())
		})

		It("delegates to underlying connection for LocalAddr", func() {
			mockAddr := &net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 8080,
			}
			mockConn := &mockNetConn{
				localAddr: mockAddr,
			}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
			}

			addr := tc.LocalAddr()

			Expect(addr).To(Equal(mockAddr))
		})

		It("delegates to underlying connection for RemoteAddr", func() {
			mockAddr := &net.TCPAddr{
				IP:   net.ParseIP("192.168.1.1"),
				Port: 5432,
			}
			mockConn := &mockNetConn{
				remoteAddr: mockAddr,
			}

			tc := &helpers.TimeoutConn{
				Conn: mockConn,
			}

			addr := tc.RemoteAddr()

			Expect(addr).To(Equal(mockAddr))
		})
	})

	Describe("BBSDBParam", func() {
		It("should have all required fields", func() {
			param := &helpers.BBSDBParam{
				DriverName:                    "postgres",
				DatabaseConnectionString:      "postgres://localhost",
				SqlCACertFile:                 "/path/to/ca.crt",
				SqlEnableIdentityVerification: true,
				ConnectionTimeout:             5 * time.Second,
				ReadTimeout:                   10 * time.Second,
				WriteTimeout:                  10 * time.Second,
			}

			Expect(param.DriverName).To(Equal("postgres"))
			Expect(param.DatabaseConnectionString).To(Equal("postgres://localhost"))
			Expect(param.SqlCACertFile).To(Equal("/path/to/ca.crt"))
			Expect(param.SqlEnableIdentityVerification).To(BeTrue())
			Expect(param.ConnectionTimeout).To(Equal(5 * time.Second))
			Expect(param.ReadTimeout).To(Equal(10 * time.Second))
			Expect(param.WriteTimeout).To(Equal(10 * time.Second))
		})
	})
})

// mockNetConn is a mock implementation of net.Conn for testing
type mockNetConn struct {
	data               []byte
	readErr            error
	writeErr           error
	closeCalled        bool
	readDeadlineSetTo  *time.Time
	writeDeadlineSetTo *time.Time
	readDeadlineErr    error
	writeDeadlineErr   error
	localAddr          net.Addr
	remoteAddr         net.Addr
}

func (m *mockNetConn) Read(b []byte) (n int, err error) {
	if m.readErr != nil {
		return 0, m.readErr
	}
	if len(m.data) == 0 {
		return 0, io.EOF
	}
	n = copy(b, m.data)
	m.data = m.data[n:]
	return n, nil
}

func (m *mockNetConn) Write(b []byte) (n int, err error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return len(b), nil
}

func (m *mockNetConn) Close() error {
	m.closeCalled = true
	return nil
}

func (m *mockNetConn) LocalAddr() net.Addr {
	return m.localAddr
}

func (m *mockNetConn) RemoteAddr() net.Addr {
	return m.remoteAddr
}

func (m *mockNetConn) SetDeadline(_ time.Time) error {
	return nil
}

func (m *mockNetConn) SetReadDeadline(t time.Time) error {
	if m.readDeadlineErr != nil {
		return m.readDeadlineErr
	}
	m.readDeadlineSetTo = &t
	return nil
}

func (m *mockNetConn) SetWriteDeadline(t time.Time) error {
	if m.writeDeadlineErr != nil {
		return m.writeDeadlineErr
	}
	m.writeDeadlineSetTo = &t
	return nil
}
