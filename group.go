package multikube

import (

)

type Group struct {
	Name string
	clusters []Cluster
}

func (g *Group) AddCluster(conf ...*Options) *Group {
	for _, c := range conf {
		g.clusters = append(g.clusters, Cluster{Options: c})
	}
	return g
}

func (g *Group) Clusters() []Cluster {
	clusters := make([]Cluster, 0)
	for _, cluster := range g.clusters {
		clusters = append(clusters, cluster)
	}
	return clusters
}

func (g *Group) SyncAll() error {
	for _, c := range g.Clusters() {
		_, err := c.SyncHTTP()
		if err != nil {
			return err
		}
	}
	return nil
}