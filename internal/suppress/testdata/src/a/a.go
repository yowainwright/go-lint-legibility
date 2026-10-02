package a

func trailing() {
	if true && false { //nolint:LEG002
		println("x")
	}
}

func ownLineAbove() {
	//nolint:legibility // reason
	if true && false {
		println("x")
	}
}

func bare() {
	if true && false { //nolint // reason
		println("x")
	}
}

func byName() {
	//nolint:hoist-if-operators,gocritic
	if true && false {
		println("x")
	}
}

func unrelatedLinter() {
	//nolint:gocritic
	if true && false { // want `LEG002`
		println("x")
	}
}

func notSuppressed() {
	if true && false { // want `LEG002`
		println("x")
	}
}

func staleAbove() {
	//nolint:LEG002

	if true && false { // want `LEG002`
		println("x")
	}
}
