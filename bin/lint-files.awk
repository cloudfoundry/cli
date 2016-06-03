!(\
       /^vendor\// \
    || /^fixtures\/.*main redeclared in this block/ \
    || /^fixtures\/.*other declaration of main/ \
    || /^plugin_examples\/.*main redeclared in this block/ \
    || /^plugin_examples\/.*other declaration of main/ \
    || /fakes\// \
    || /cf\/resources.*\(golint\)/ \
    || /words\/.*\(golint\)/ \
    || /plugin\/.*\(golint\)/ \
    || /cf\/resources.*\(gofmt\)/ \
    || /words\/.*\(gofmt\)/ \
)