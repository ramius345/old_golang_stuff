FROM busybox:musl
ADD cleanup /
ADD filescan /
ADD thumbnailer
CMD [ "/bin/sh", "-c", "echo Running Cleanup; /cleanup; echo Running Filescan; /filescan; echo Running thumbnailer; /thumbnailer"]
