package zeroinit_test

import (
	"context"
	"fmt"

	. "github.com/mudler/zeroinit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("zeroinit dag", func() {
	var g *Graph
	BeforeEach(func() {
		//	EventuallyConnects(1200)
		g = NewGraph()
	})

	Context("simple checks", func() {
		It("orders", func() {
			g.DependOn("A", "B")
			g.DependOn("B", "C")
			g.DependOn("C", "D")
			g.DependOn("D", "E")
			Expect(g.TopoSortedLayers()).To(Equal([][]string{[]string{"E"}, []string{"D"}, []string{"C"}, []string{"B"}, []string{"A"}}))
		})

		It("orders parallel", func() {
			g.DependOn("A", "B")
			g.DependOn("B", "C")
			g.DependOn("C", "D")
			g.DependOn("D", "E")
			g.DependOn("X", "E")
			Expect(g.TopoSortedLayers()).To(
				Or(
					Equal([][]string{[]string{"E"}, []string{"D", "X"}, []string{"C"}, []string{"B"}, []string{"A"}}),
					Equal([][]string{[]string{"E"}, []string{"X", "D"}, []string{"C"}, []string{"B"}, []string{"A"}}),
				),
			)
		})
	})

	Context("Sequential runs", func() {
		It("orders parallel", func() {
			f := ""
			g.AddOp("foo", WithCallback(func(ctx context.Context) error {
				f += "foo"
				return nil
			}), WithDeps("bar"))
			g.AddOp("bar", WithCallback(func(ctx context.Context) error {
				f += "bar"
				return nil
			}))
			g.Run(context.Background())
			Expect(f).To(Equal("barfoo"))
		})
	})

	Context("With errors", func() {
		It("fails", func() {
			f := ""

			g.AddOp("foo", WithCallback(func(ctx context.Context) error {
				return fmt.Errorf("failure")
			}), WithDeps("bar"), FatalOp)

			g.AddOp("bar",
				WithCallback(func(ctx context.Context) error {
					f += "bar"
					return nil
				}),
			)

			err := g.Run(context.Background())
			Expect(err).To(Equal(fmt.Errorf("failure")))
		})
	})

	Context("Sequential runs, background jobs", func() {
		It("orders parallel", func() {
			testChan := make(chan string)
			f := ""
			g.AddOp("foo", WithCallback(func(ctx context.Context) error {
				f += "triggered"
				return nil
			}), WithDeps("bar"))
			g.AddOp("bar", WithCallback(func(ctx context.Context) error {
				<-testChan
				return fmt.Errorf("test")
			}), Background)
			g.Run(context.Background())
			Expect(g.State("bar").Error).ToNot(HaveOccurred())
			Expect(f).To(Equal("triggered"))
			testChan <- "foo"
			Expect(g.State("bar").Error).To(HaveOccurred())
		})
	})
})
