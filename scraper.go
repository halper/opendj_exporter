package opendj_exporter

import (
	"context"
	"fmt"
	ldap "github.com/go-ldap/ldap/v3"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

const (
	namespace = "opendj"
)

var (
	baseMonitor                  = "cn=monitor"
	administrationConnectorGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "administration_connector",
			Help:      "Administration connector",
		},
		[]string{"attr"},
	)
	backendGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "userRoot_backend",
			Help:      "userRoot backend",
		},
		[]string{"attr"},
	)
	memoryGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "jvm_memory_usage",
			Help:      "JVM memory usage",
		},
		[]string{"attr"},
	)
	ldapGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ldap_connection_handler",
			Help:      "LDAP connection handler metrics",
		},
		[]string{"attr"},
	)
	ldapsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ldaps_connection_handler",
			Help:      "LDAPS connection handler metrics",
		},
		[]string{"attr"},
	)
	databaseGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "userRoot_je_database",
			Help:      "userRoot JE database",
		},
		[]string{"attr"},
	)
	administrationStatisticsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "administration_connector_statistics",
			Help:      "Administration connector statistics",
		},
		[]string{"attr"},
	)
	ldapHandlerGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ldap_connection_handler_statistics",
			Help:      "LDAP connection handler statistics",
		},
		[]string{"attr"},
	)
	ldapsHandlerGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "ldaps_connection_handler_statistics",
			Help:      "LDAPS connection handler statistics",
		},
		[]string{"attr"},
	)
	queries []*query
)

type Scraper struct {
	Addr                    string
	User                    string
	Pass                    string
	Tick                    time.Duration
	log                     log.FieldLogger
	Conn                    *ldap.Conn
	LdapListenAddr          string
	LdapsListenAddr         string
	AdministrationConnector string
	LdapPort                int
	LdapsPort               int
	AdministrationPort      int
}

type query struct {
	baseDN  string
	metric  *prometheus.GaugeVec
	setData func([]*ldap.Entry, *query)
}

func getBaseDN(cn string) string {
	return fmt.Sprintf("cn=%s,%s", cn, baseMonitor)
}

func buildQueries(s *Scraper) {
	queries = append(queries, &query{
		baseDN:  getBaseDN("userRoot JE Database"),
		metric:  databaseGauge,
		setData: setValue,
	})
	queries = append(queries, &query{
		baseDN:  getBaseDN("JVM Memory Usage"),
		metric:  memoryGauge,
		setData: setValue,
	})
	queries = append(queries, &query{
		baseDN:  getBaseDN("userRoot backend"),
		metric:  backendGauge,
		setData: setValue,
	})
	builtCn := fmt.Sprintf("Administration Connector %s port %d", s.AdministrationConnector, s.AdministrationPort)
	queries = append(queries, &query{
		baseDN:  getBaseDN(builtCn),
		metric:  administrationConnectorGauge,
		setData: setValue,
	})
	builtCn = fmt.Sprintf("LDAP Connection Handler %s port %d", s.LdapListenAddr, s.LdapPort)
	queries = append(queries, &query{
		baseDN:  getBaseDN(builtCn),
		metric:  ldapGauge,
		setData: setValue,
	})
	builtCn = fmt.Sprintf("LDAPS Connection Handler %s port %d", s.LdapsListenAddr, s.LdapsPort)
	queries = append(queries, &query{
		baseDN:  getBaseDN(builtCn),
		metric:  ldapsGauge,
		setData: setValue,
	})
	builtCn = fmt.Sprintf("Administration Connector %s port %d Statistics", s.AdministrationConnector, s.AdministrationPort)
	queries = append(queries, &query{
		baseDN:  getBaseDN(builtCn),
		metric:  administrationStatisticsGauge,
		setData: setValue,
	})
	builtCn = fmt.Sprintf("LDAP Connection Handler %s port %d Statistics", s.LdapListenAddr, s.LdapPort)
	queries = append(queries, &query{
		baseDN:  getBaseDN(builtCn),
		metric:  ldapHandlerGauge,
		setData: setValue,
	})
	builtCn = fmt.Sprintf("LDAPS Connection Handler %s port %d Statistics", s.LdapsListenAddr, s.LdapsPort)
	queries = append(queries, &query{
		baseDN:  getBaseDN(builtCn),
		metric:  ldapsHandlerGauge,
		setData: setValue,
	})
}

func setValue(entries []*ldap.Entry, q *query) {
	for _, entry := range entries {
		for _, attribute := range entry.Attributes {
			attrName := attribute.Name
			attrVal := attribute.Values[0]
			if n, err := strconv.ParseFloat(attrVal, 64); err == nil {
				q.metric.WithLabelValues(attrName).Set(n)
			}
		}
	}
}
func init() {
	prometheus.MustRegister(
		databaseGauge,
		memoryGauge,
		backendGauge,
		ldapsGauge,
		ldapGauge,
		administrationConnectorGauge,
		administrationStatisticsGauge,
		ldapsHandlerGauge,
		ldapHandlerGauge,
	)
}
func (s *Scraper) Start(ctx context.Context) {
	var err error
	s.log = log.WithField("component", "scraper")
	buildQueries(s)

	address := fmt.Sprintf("tcp://%s", s.Addr)
	s.log.WithField("addr", address).Info("starting monitor loop")
	s.Conn, err = ldap.Dial("tcp", s.Addr)
	if err != nil {
		s.log.WithError(err).Error("dial failed")
		return
	}
	ticker := time.NewTicker(s.Tick)
	defer ticker.Stop()
	defer s.Conn.Close()
	for {
		select {
		case <-ticker.C:
			s.scrape()
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scraper) scrape() {
	var err error
	conn := s.Conn
	if s.User != "" && s.Pass != "" {
		err = conn.Bind(s.User, s.Pass)
		if err != nil {
			s.log.WithError(err).Error("bind failed")
			return
		}
	}

	for _, q := range queries {
		if err = scrapeQuery(conn, q); err != nil {
			s.log.WithError(err).WithField("name", q.baseDN).Warn("query failed")
		}
	}
}

func scrapeQuery(conn *ldap.Conn, q *query) error {
	req := ldap.NewSearchRequest(
		q.baseDN, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectclass=*)", []string{"*"}, nil,
	)
	sr, err := conn.Search(req)
	if err != nil {
		return err
	}
	q.setData(sr.Entries, q)
	return nil
}
