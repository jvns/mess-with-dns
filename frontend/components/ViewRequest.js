import template from "./ViewRequest.html";

export default {
    template: template,
    props: ["log"],
    data: function() {
        return {
            clicked: false,
        };
    },
    methods: {
        localTime: function(timestamp) {
            // convert to local time
            const date = new Date(timestamp * 1000);
            // only show time, not day
            // 24-hour format
            const hours = date.getHours();
            const minutes = "0" + date.getMinutes();
            const seconds = "0" + date.getSeconds();
            const formattedTime = hours + ":" + minutes.substr(-2) + ":" + seconds.substr(-2);
            return formattedTime;
        },
    },
};
