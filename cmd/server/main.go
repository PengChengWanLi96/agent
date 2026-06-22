package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"agent/internal/api"
	"agent/internal/client/docker"
	"agent/internal/client/metrics"
	"agent/internal/config"
	"agent/internal/service"
	"agent/internal/version"
)

func main() {
	cfg := config.Load()

	var (
		showVersion bool
		port        int
		addr        string
	)

	fs := flag.NewFlagSet("agent", flag.ExitOnError)
	fs.BoolVar(&showVersion, "version", false, "打印版本信息并退出")
	fs.BoolVar(&showVersion, "v", false, "打印版本信息并退出（简写）")
	fs.IntVar(&port, "port", 0, "HTTP 服务监听端口（覆盖 SERVER_ADDR 中的端口）")
	fs.IntVar(&port, "p", 0, "HTTP 服务监听端口（简写）")
	fs.StringVar(&addr, "addr", "", "HTTP 服务监听地址，如 :8080 或 0.0.0.0:8080（覆盖 SERVER_ADDR）")
	fs.StringVar(&cfg.Docker.Host, "docker-host", cfg.Docker.Host, "Docker 守护进程地址")
	fs.StringVar(&cfg.Docker.APIVersion, "docker-api-version", cfg.Docker.APIVersion, "Docker API 版本")
	fs.BoolVar(&cfg.Docker.TLSVerify, "docker-tls-verify", cfg.Docker.TLSVerify, "启用 Docker TLS 验证")
	fs.StringVar(&cfg.Docker.CertPath, "docker-cert-path", cfg.Docker.CertPath, "Docker TLS 证书目录路径")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "agent - 服务管理 Agent\n\n")
		fmt.Fprintf(os.Stderr, "用法:\n  agent [选项]\n\n选项:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n环境变量:\n")
		fmt.Fprintf(os.Stderr, "  SERVER_ADDR, DOCKER_HOST, DOCKER_API_VERSION, DOCKER_TLS_VERIFY, DOCKER_CERT_PATH\n")
		fmt.Fprintf(os.Stderr, "  （命令行选项优先级高于环境变量）\n\n")
		fmt.Fprintf(os.Stderr, "示例:\n")
		fmt.Fprintf(os.Stderr, "  agent --port 9090\n")
		fmt.Fprintf(os.Stderr, "  agent --addr 0.0.0.0:8080 --docker-host tcp://192.168.1.100:2375\n")
		fmt.Fprintf(os.Stderr, "  agent --version\n")
	}

	if err := fs.Parse(os.Args[1:]); err != nil {
		os.Exit(2)
	}

	if showVersion {
		fmt.Println(version.String())
		return
	}

	// 命令行选项优先级高于环境变量
	if addr != "" {
		cfg.ServerAddr = addr
	}
	if port != 0 {
		if port < 1 || port > 65535 {
			log.Fatalf("invalid port: %d (必须在 1-65535 范围内)", port)
		}
		cfg.ServerAddr = ":" + strconv.Itoa(port)
	}

	dockerCli, err := docker.NewClient(cfg.Docker.Host, cfg.Docker.TLSVerify, cfg.Docker.CertPath)
	if err != nil {
		log.Fatalf("failed to create docker client: %v", err)
	}
	defer dockerCli.Close()

	if err := dockerCli.Ping(context.Background()); err != nil {
		log.Printf("docker ping failed: %v", err)
	}

	metricsCollector, err := metrics.NewCollector()
	if err != nil {
		log.Fatalf("failed to create metrics collector: %v", err)
	}

	dockerSvc := service.NewDockerService(dockerCli)
	metricsSvc := service.NewMetricsService(metricsCollector, time.Now())
	sshSvc := service.NewSSHService()

	r := api.NewRouter(dockerSvc, metricsSvc, sshSvc)

	srv := &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}

	go func() {
		log.Printf("agent %s starting on %s", version.Short(), cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server exited")
}
