package pkg

import (
	birelpkg "github.com/cloudfoundry/bosh-cli/release/pkg"
)

func ResolveDependencies(pkg birelpkg.Compilable) []birelpkg.Compilable {
	return reverse(resolveInner(pkg, []birelpkg.Compilable{}))
}

func resolveInner(pkg birelpkg.Compilable, noFollow []birelpkg.Compilable) []birelpkg.Compilable {
	all := []birelpkg.Compilable{}

	for _, depPkg := range pkg.Deps() {
		if !contains(all, depPkg) && !contains(noFollow, depPkg) {
			all = append(all, depPkg)

			tDeps := resolveInner(depPkg, joinUnique(all, noFollow))
			for _, tDepPkg := range tDeps {
				all = append(all, tDepPkg)
			}
		}
	}

	for i, el := range all {
		if el == pkg {
			all = append(all[:i], all[i+1:]...)
		}
	}

	return all
}

func contains(list []birelpkg.Compilable, element birelpkg.Compilable) bool {
	for _, pkg := range list {
		if element == pkg {
			return true
		}
	}
	return false
}

func joinUnique(a []birelpkg.Compilable, b []birelpkg.Compilable) []birelpkg.Compilable {
	joined := []birelpkg.Compilable{}
	joined = append(joined, a...)
	for _, pkg := range b {
		if !contains(a, pkg) {
			joined = append(joined, pkg)
		}
	}
	return joined
}

func reverse(a []birelpkg.Compilable) []birelpkg.Compilable {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}
