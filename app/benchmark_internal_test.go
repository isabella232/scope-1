package app

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

var (
	benchReportPath = flag.String("bench-report-path", "", "report file, or dir with files, to use for benchmarking (relative to this package)")
)

func readReportFiles(b *testing.B, path string) []report.Report {
	reports := []report.Report{}
	if err := filepath.Walk(path,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rpt, err := report.MakeFromFile(p)
			if err != nil {
				return err
			}
			reports = append(reports, rpt)
			return nil
		}); err != nil {
		b.Fatal(err)
	}
	return reports
}

func BenchmarkReportUnmarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		readReportFiles(b, *benchReportPath)
	}
}

func BenchmarkReportUpgrade(b *testing.B) {
	reports := readReportFiles(b, *benchReportPath)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, r := range reports {
			r.Upgrade()
		}
	}
}

func BenchmarkReportMerge(b *testing.B) {
	reports := readReportFiles(b, *benchReportPath)
	merger := NewSmartMerger()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		merger.Merge(reports)
	}
}

func benchmarkRender(b *testing.B, f func(report.Report)) {
	r := fixture.Report
	if *benchReportPath != "" {
		r = NewSmartMerger().Merge(readReportFiles(b, *benchReportPath))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		render.ResetCache()
		b.StartTimer()
		f(r)
	}
}

func benchmarkRenderTopology(b *testing.B, topologyID string) {
	benchmarkRender(b, func(report report.Report) {
		renderer, filter, err := topologyRegistry.RendererForTopology(topologyID, url.Values{}, report)
		if err != nil {
			b.Fatal(err)
		}
		render.Render(report, renderer, filter)
	})
}

func BenchmarkRenderList(b *testing.B) {
	benchmarkRender(b, func(report report.Report) {
		request := &http.Request{
			Form: url.Values{},
		}
		topologyRegistry.renderTopologies(report, request)
	})
}

func BenchmarkRenderHosts(b *testing.B) {
	benchmarkRenderTopology(b, "hosts")
}

func BenchmarkRenderControllers(b *testing.B) {
	benchmarkRenderTopology(b, "kube-controllers")
}

func BenchmarkRenderPods(b *testing.B) {
	benchmarkRenderTopology(b, "pods")
}

func BenchmarkRenderContainers(b *testing.B) {
	benchmarkRenderTopology(b, "containers")
}

func BenchmarkRenderProcesses(b *testing.B) {
	benchmarkRenderTopology(b, "processes")
}
