package pkg

import (
	"errors"
	"fmt"
)

// Topologically sorts an array of packages
func Sort(releasePackages []Compilable) ([]Compilable, error) {
	sortedPackages := []Compilable{}

	incomingEdges, outgoingEdges := getEdgeMaps(releasePackages)
	noIncomingEdgesSet := []Compilable{}

	for pkg, edgeList := range incomingEdges {
		if len(edgeList) == 0 {
			noIncomingEdgesSet = append(noIncomingEdgesSet, pkg)
		}
	}
	for len(noIncomingEdgesSet) > 0 {
		elem := noIncomingEdgesSet[0]
		noIncomingEdgesSet = noIncomingEdgesSet[1:]

		sortedPackages = append([]Compilable{elem}, sortedPackages...)

		for _, pkg := range outgoingEdges[elem] {
			incomingEdges[pkg] = removeFromList(incomingEdges[pkg], elem)
			if len(incomingEdges[pkg]) == 0 {
				noIncomingEdgesSet = append(noIncomingEdgesSet, pkg)
			}
		}
	}
	for _, edges := range incomingEdges {
		if len(edges) > 0 {
			return nil, errors.New("Circular dependency detected while sorting packages")
		}
	}
	return sortedPackages, nil
}

func removeFromList(packageList []Compilable, pkg Compilable) []Compilable {
	for idx, elem := range packageList {
		if elem == pkg {
			return append(packageList[:idx], packageList[idx+1:]...)
		}
	}
	panic(fmt.Sprintf("Expected %s to be in dependency graph", pkg.Name()))
}

func getEdgeMaps(releasePackages []Compilable) (map[Compilable][]Compilable, map[Compilable][]Compilable) {
	incomingEdges := make(map[Compilable][]Compilable)
	outgoingEdges := make(map[Compilable][]Compilable)

	for _, pkg := range releasePackages {
		incomingEdges[pkg] = []Compilable{}
	}

	for _, pkg := range releasePackages {
		if pkg.Deps() != nil {
			for _, dep := range pkg.Deps() {
				incomingEdges[dep] = append(incomingEdges[dep], pkg)
				outgoingEdges[pkg] = append(outgoingEdges[pkg], dep)
			}
		}
	}
	return incomingEdges, outgoingEdges
}
