package pool

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

type Pool struct {
	maxCnt      int
	maxIdleTime time.Duration
	factory     func() (net.Conn, error)

	mu        sync.Mutex
	idleConns chan *idleConn
	// current conns count
	cnt      int
	connReqs []connReq
}

func NewPool(cfg *PoolConfig) (*Pool, error) {
	if cfg.InitCnt > cfg.MaxIdleCnt {
		return nil, errors.New("init cnt can't be bigger than max cnt")
	}

	idleConns := make(chan *idleConn, cfg.MaxIdleCnt)
	for i := 0; i < cfg.InitCnt; i++ {
		conn, err := cfg.Factory()
		if err != nil {
			return nil, err
		}
		idleConns <- &idleConn{
			c:              conn,
			lastActiveTime: time.Now(),
		}
	}
	pool := &Pool{
		idleConns:   idleConns,
		maxCnt:      cfg.MaxCnt,
		maxIdleTime: cfg.MaxIdleTime,
		factory:     cfg.Factory,
	}
	return pool, nil
}

type PoolConfig struct {
	InitCnt     int
	MaxCnt      int
	MaxIdleCnt  int
	MaxIdleTime time.Duration
	Factory     func() (net.Conn, error)
}

type idleConn struct {
	c              net.Conn
	lastActiveTime time.Time
}

type connReq struct {
	conn chan net.Conn
}

func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
L:
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case idleConn := <-p.idleConns:
			if idleConn.lastActiveTime.Add(p.maxIdleTime).Before(time.Now()) {
				_ = idleConn.c.Close()
				continue
			}
			return idleConn.c, nil
		default:
			break L
		}
	}

	// no idle conn
	p.mu.Lock()
	if p.cnt >= p.maxCnt {
		req := connReq{
			conn: make(chan net.Conn, 1),
		}
		p.connReqs = append(p.connReqs, req)
		p.mu.Unlock()
		select {
		case <-ctx.Done():
			go func() {
				c := <-req.conn
				_ = p.Put(context.Background(), c)
			}()
			return nil, ctx.Err()
		case c := <-req.conn:
			return c, nil
		}
	}

	c, err := p.factory()
	if err != nil {
		return nil, err
	}
	p.cnt++
	p.mu.Unlock()
	return c, nil
}

func (p *Pool) Put(ctx context.Context, c net.Conn) error {
	p.mu.Lock()

	if len(p.connReqs) > 0 {
		req := p.connReqs[len(p.connReqs)-1]
		p.connReqs = p.connReqs[:len(p.connReqs)-1]
		p.mu.Unlock()
		req.conn <- c
		return nil
	}

	defer p.mu.Unlock()

	idleConn := idleConn{
		c:              c,
		lastActiveTime: time.Now(),
	}
	select {
	case p.idleConns <- &idleConn:
	default:
		c.Close()
		p.cnt--
	}

	return nil
}
