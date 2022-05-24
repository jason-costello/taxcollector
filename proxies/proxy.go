package proxies

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/jason-costello/taxcollector/storage/pgdb"
	_ "github.com/mattn/go-sqlite3"
)

type ProxyClient struct {
	hc  *http.Client
	pdb *pgdb.Queries
	db  *sql.DB
}

func NewProxyClient(db *sql.DB) *ProxyClient {

	return &ProxyClient{
		db:  db,
		pdb: pgdb.New(db),
	}

}

type Proxy struct {
	IP       string    `json:"IP"`
	LastUsed time.Time `json:"lastUsed"`
	Uses     int       `json:"uses"`
	IsBad    bool      `json:"is_bad"`
}

func (p *ProxyClient) GetNext() (Proxy, error) {

	proxyRow, err := p.pdb.GetValidProxy(context.Background())
	if err != nil {
		return Proxy{}, err
	}

	if proxyRow.Ip == "" {
		return Proxy{}, errors.New("no proxy ip found")
	}

	var uses int32 = 0
	if proxyRow.Uses.Valid {
		uses = proxyRow.Uses.Int32
	}

	proxy := Proxy{
		IP:   proxyRow.Ip,
		Uses: int(uses),
	}

	if err := p.UpdateLastUsed(&proxy); err != nil {
		return Proxy{}, err
	}

	return proxy, nil

}

func (p *ProxyClient) UpdateLastUsed(proxy *Proxy) error {
	if proxy.IP == "" {
		return errors.New("no IP provided")
	}

	params := pgdb.UpdateProxyLastUsedTimeParams{
		Lastused: sql.NullString{String: time.Now().String(), Valid: true},
		Uses:     sql.NullInt32{Int32: int32(proxy.Uses + 1), Valid: true},
		Ip:       proxy.IP,
	}
	if err := p.pdb.UpdateProxyLastUsedTime(context.Background(), params); err != nil {
		return err
	}

	return nil

}

func (p *ProxyClient) MarkProxyAsBad(proxyIP string) error {
	updateQuery := `update proxies set is_bad = true where ip = ?`
	updateStmt, err := p.db.Prepare(updateQuery)
	if err != nil {
		return err
	}

	_, err = updateStmt.Exec(proxyIP)
	if err != nil {
		return err
	}
	return nil

}
