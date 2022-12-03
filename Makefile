build:
	cargo build

test:
	cargo run -q -- --config ./aur/config.json --log-level=trace --port=1053

fmt:
	cargo fmt
