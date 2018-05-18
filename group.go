package multikube

type Group struct {
	Name string
	clusters map[string]Cluster
}

func (g *Group) AddCluster(conf ...*ClusterConfig) *Group {
	for _, c := range conf {
		g.clusters[c.Name] = Cluster{ Config: c }
	}
	return g
}

func (g *Group) Cluster(name string) *Cluster {
	cluster := g.clusters[name]
	return &cluster
}

func (g *Group) Clusters() []Cluster {
	clusters := make([]Cluster, 0)
	for _, cluster := range g.clusters {
		clusters = append(clusters, cluster)
	}
	return clusters
}