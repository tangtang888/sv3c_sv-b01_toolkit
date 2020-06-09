with import <nixpkgs> {};
stdenv.mkDerivation rec {
	name = "onvif_record";
	env = buildEnv { name = name; paths = buildInputs; };
	
	buildInputs = [
		go
	];
}
