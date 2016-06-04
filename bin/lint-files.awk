!(\
       /^vendor\// \
    || /fakes\// \
    || /^fixtures\/.*main redeclared in this block/ \
    || /^fixtures\/.*other declaration of main/ \
    || /^plugin_examples\/.*main redeclared in this block/ \
    || /^plugin_examples\/.*other declaration of main/ \
    || /cf\/resources.*\(golint\)/ \
    || /words\/.*\(golint\)/ \
    || /plugin\/.*\(golint\)/ \
    || /cf\/resources.*\(gofmt\)/ \
    || /words\/.*\(gofmt\)/ \
    || /_test\.go.*\(errcheck\)/ \
)