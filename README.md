# pathsort-go

A small utility that orders the elements in your PATH according to a TOML config file. It reads the config file, which defines regular expressions that will be matched against each element of your PATH. If there's a match, the name associated with the regular expression is used to place that PATH element in an order defined by the `tag_order` array in the config. Unmatched path elements are added to the end of the ordered PATH.
