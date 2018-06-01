package multikube

import (

)

type Group struct {
	Name string
	clusters []Cluster
}

func (g *Group) AddCluster(conf ...*Config) *Group {
	for _, c := range conf {
		g.clusters[c.Name] = Cluster{ Config: c }
	}
	return g
}

func (g *Group) Cluster(name string) *Cluster {
	for _, c := range g.clusters {
		if c.Name == name {
			return &c
		}
	}
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