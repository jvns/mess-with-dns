import {
    rrTypesReverse
} from "../common.js";
import template from "./ViewRequest.html";
export default {
    template: template,
    props: ["event"],
    data: function() {
        return {
            clicked: false,
            rrTypesReverse: rrTypesReverse,
        };
    },
    methods: {
        localTime: function(timestamp) {
            // convert to local time
            var date = new Date(timestamp * 1000);
            // only show time, not day
            // 24-hour format
            var hours = date.getHours();
            var minutes = "0" + date.getMinutes();
            var seconds = "0" + date.getSeconds();
            var formattedTime = hours + ":" + minutes.substr(-2) + ":" + seconds.substr(-2);
            return formattedTime;
        },
        formatFull(answer) {
            return rrTypesReverse[answer.Hdr.Rrtype] + " " + this.formatContent(answer);
        },
        formatContent(answer) {
            var result = "";
            for (var key in answer) {
                if (key == "Hdr") {
                    continue;
                }
                result += answer[key] + " ";
            }
            return result;
        },
    },
};
