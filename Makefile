.PHONY: dist.tar.gz
dist.tar.gz: dist
	rm --force $@
	tar --create --gzip --file $@ $<


.PHONY: dist
dist:
	rm --recursive --force $@
	DOCKER_BUILDKIT=1 docker build \
		--build-arg BUILDKIT_INLINE_CACHE=1 \
		--output $@ \
		.
