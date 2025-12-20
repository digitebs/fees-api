package bill

TemporalServer: [
	// These act as individual case statements
    if #Meta.Environment.Cloud == "local" { "localhost:7233" },

    // TODO: configure this to match your own cluster address
    "localhost:7233",
][0] // Return the first value which matches the condition
