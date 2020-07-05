with import <nixpkgs> {};
stdenv.mkDerivation rec {
	name = "sv3c_b01_onvif_mqtt";
	env = buildEnv { name = name; paths = buildInputs; };
	
	buildInputs = [
		go
	];
}
