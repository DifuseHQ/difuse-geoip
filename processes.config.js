module.exports = {
    apps : [{
        name: "difuse-geoip",
        script: "./difuse-geoip",
        watch: false,
        instances: 1,
        exec_mode: "fork",
    }]
};
