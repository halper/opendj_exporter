package main

import (
  "os/signal"
	"syscall"
  "golang.org/x/sync/errgroup"
  exporter "github.com/halper/opendj_exporter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
  "context"
	"os"
	"sort"
	"time"
  log "github.com/sirupsen/logrus"
)

const (
	promAddr          = "promAddr"
	ldapAddr          = "ldapAddr"
	ldapUser          = "ldapUser"
	ldapPass          = "ldapPass"
	interval          = "interval"
	metrics           = "metrics"
	configFile            = "configFile"
  ldapPort = "ldapPort"
  ldapListenAddr = "ldapListenAddr"
  ldapsPort = "ldapsPort"
  ldapsListenAddr = "ldapsListenAddr"
  adminPort = "adminPort"
  adminListenAddr = "adminListenAddr"
  webCfgFile = "webCfgFile"
)

func main() {
	// define flags
	flags := []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    promAddr,
			Aliases: []string{"a"},
			Value:   ":9330",
			Usage:   "Bind address for Prometheus HTTP metrics server",
			EnvVars: []string{"PROM_ADDR"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    metrics,
			Aliases: []string{"m"},
			Value:   "/metrics",
			Usage:   "Path on which to expose Prometheus metrics",
			EnvVars: []string{"METRICS_PATH"},
		}),
    altsrc.NewStringFlag(&cli.StringFlag{
			Name:    webCfgFile,
			Usage:   "Prometheus metrics web config `FILE` (optional)",
			EnvVars: []string{"WEB_CFG_FILE"},
		}),
		altsrc.NewDurationFlag(&cli.DurationFlag{
			Name:    interval,
			Aliases: []string{"i"},
			Value:   30 * time.Second,
			Usage:   "Scrape interval",
			EnvVars: []string{"INTERVAL"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    ldapAddr,
			Aliases: []string{"l"},
			Value:   "localhost:389",
			Usage:   "Address and port of OpenDJ server",
			EnvVars: []string{"LDAP_ADDR"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    ldapPort,
      Value:   "389",
			Usage:   "OpenDJ LDAP port",
			EnvVars: []string{"LDAP_PORT"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    ldapListenAddr,
      Value:   "0.0.0.0",
			Usage:   "The address that LDAP connection handler is listening",
			EnvVars: []string{"LDAP_LSTN"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    ldapsPort,
      Value:   "636",
			Usage:   "OpenDJ LDAPS port",
			EnvVars: []string{"LDAPS_PORT"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    ldapsListenAddr,
      Value:   "0.0.0.0",
			Usage:   "The address that LDAPS connection handler is listening",
			EnvVars: []string{"LDAPS_LSTN"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    adminPort,
      Value:   "4444",
			Usage:   "OpenDJ Administration port",
			EnvVars: []string{"ADMN_PORT"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    adminListenAddr,
      Value:   "0.0.0.0",
			Usage:   "The address that administration connector is listening",
			EnvVars: []string{"ADMN_LSTN"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    ldapUser,
			Aliases: []string{"u"},
			Usage:   "OpenDJ bind username (optional)",
			EnvVars: []string{"LDAP_USER"},
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    ldapPass,
			Usage:   "OpenDJ bind password (optional)",
			EnvVars: []string{"LDAP_PASS"},
		}),
    &cli.StringFlag{
			Name:    configFile,
			Aliases: []string{"c"},
			Usage:   "Optional configuration from a `YAML_FILE`",
		},
	}
	// define app
	app := &cli.App{
		Name:   "opendj_exporter",
		Usage:  "Export OpenDJ metrics to Prometheus",
		Before: altsrc.InitInputSourceWithContext(flags, optionalYamlSourceFunc("configFile")),
    Version:         exporter.GetVersion(),
		HideHelpCommand: true,
		Flags:           flags,
		Action:          runMain,
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	if err := app.Run(os.Args); err != nil {
		log.WithError(err).Fatal("service failed")
	}
	log.Info("service stopped")
}

func optionalYamlSourceFunc(flagFileName string) func(context *cli.Context) (altsrc.InputSourceContext, error) {
	return func(c *cli.Context) (altsrc.InputSourceContext, error) {
		filePath := c.String(flagFileName)
		if _, err := os.Stat(filePath); err == nil {
			return altsrc.NewYamlSourceFromFile(filePath)
		}
		return &altsrc.MapInputSource{}, nil
	}
}

func runMain(c *cli.Context) error {
	server := exporter.NewMetricsServer(
		c.String(promAddr),
		c.String(metrics),
    c.String(webCfgFile),
	)

	scraper := &exporter.Scraper{
		Addr: c.String(ldapAddr),
		User: c.String(ldapUser),
		Pass: c.String(ldapPass),
		Tick: c.Duration(interval),
    LdapListenAddr: c.String(ldapListenAddr),
    LdapsListenAddr: c.String(ldapsListenAddr),
    LdapPort: c.String(ldapPort),
    LdapsPort: c.String(ldapsPort),
    AdministrationConnector: c.String(adminListenAddr),
    AdministrationPort: c.String(adminPort),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var group errgroup.Group
	group.Go(func() error {
		defer cancel()
		return server.Start()
	})
	group.Go(func() error {
		defer cancel()
		scraper.Start(ctx)
		return nil
	})
	group.Go(func() error {
		defer func() {
			cancel()
			server.Stop()
		}()
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-signalChan:
			log.Info("shutdown received")
			return nil
		case <-ctx.Done():
			return nil
		}
	})
	return group.Wait()
}

