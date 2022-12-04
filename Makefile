build:
	cargo build
	cargo clippy

dev:
	cargo run -- --config ./aur/config.json --log-level=trace --port=1053 | jq

fmt:
	cargo fmt
