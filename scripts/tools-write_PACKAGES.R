#!/usr/bin/env Rscript

require(tools)
require(jsonlite)
require(logging)


basicConfig()

json <- file("stdin")

write_PACKAGES <- function(json){
  oldPath <- getwd()
  obj <- fromJSON(json)
  obj$targetRepo <- normalizePath(obj$targetRepo)
  loginfo("Changing to dir: %s", obj$targetRepo)
  setwd(obj$targetRepo)

  tools::write_PACKAGES(
    verbose = TRUE,
    unpacked = FALSE,
    latestOnly = FALSE,
    addFiles = TRUE
  )

  loginfo(obj)

  setwd(oldPath)
}

write_PACKAGES(json)
