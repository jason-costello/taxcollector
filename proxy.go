package main

import (
	"database/sql"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type ProxyClient struct {
	hc *http.Client
	db *sql.DB
}

func NewProxyClient(db *sql.DB) *ProxyClient {

	return &ProxyClient{
		db: db,
	}

}

type Proxy struct {
	IP       string    `json:"IP"`
	LastUsed time.Time `json:"lastUsed"`
	Uses     int       `json:"uses"`
	IsBad    bool      `json:"isBad"`
}

func (p *ProxyClient) GetNext() (Proxy, error) {

	query := `select ip, lastused, uses
		from proxies
		where isBad = false
		order by lastused asc, uses asc
		limit 1;`

	stmt, err := p.db.Prepare(query)
	if err != nil {
		return Proxy{}, err
	}
	row := stmt.QueryRow()
	var ip string
	var lastused time.Time
	var uses int

	row.Scan(&ip, &lastused, &uses)

	proxy := Proxy{
		IP:       ip,
		LastUsed: lastused,
		Uses:     uses,
	}

	if err := p.UpdateLastUsed(&proxy); err != nil {
		return Proxy{}, err
	}

	return proxy, nil

}

func (p *ProxyClient) UpdateLastUsed(proxy *Proxy) error {
	updateQuery := `update proxies set lastused = ?, uses = ? where ip = ?`
	updateStmt, err := p.db.Prepare(updateQuery)
	if err != nil {
		return err
	}
	proxy.Uses += 1
	_, err = updateStmt.Exec(time.Now(), proxy.Uses, proxy.IP)
	if err != nil {
		return err
	}
	return nil

}