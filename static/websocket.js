$(function () {
    'use strict';

    class oneManager {

        constructor(number, id, row) {
            this.number = number;
            this.id = id;
            this.parentRow = row;

            this.numberTd = $('<td></td>').attr({
                id: this.id + "-number",
                name: "number",
            }).html(this.number);

            this.nameTd = $('<td></td>').attr({
                id: this.id + "-name",
                name: "name",
            });

            this.chanBeforeTd = $('<td></td>').attr({
                id: this.id + "-before",
                name: "before",
            });
            this.chanAfterTd = $('<td></td>').attr({
                id: this.id + "-after",
                name: "after",
            });
            this.workersTd = $('<td></td>').attr({
                id: this.id + "-workers",
                name: "workers",
            });

            this.row = $('<tr></tr>').attr({
                id: this.id + "-tr",
                name: this.id,
            }).append(this.numberTd, this.nameTd, this.chanBeforeTd, this.chanAfterTd, this.workersTd)

            this.parentRow.append(this.row)
            return this
        }


        draw(val) {
            this.nameTd.html('<b>' + val.Name + '</b>');

            let b = val.ChanBefore;
            if (Array.isArray(b) && b.length > 0) {
                this.chanBeforeTd.text(b[0].IsExisted + " / " + b[0].Length + " / " + b[0].Type);
            }

            let a = val.ChanAfter;
            if (Array.isArray(a) && a.length > 0) {
                this.chanAfterTd.text(a[0].IsExisted + " / " + a[0].Length + " / " + a[0].Type);
            }
            let w = val.Workers;
            this.workersTd.text((w.Min || 0) + " / " + (w.Max || 0) + " / " + (w.Number || 0) + " / " + (w.Active || 0));
        }
    }

    //
    // <span id="alldata">
    //
    //     <div class="table-responsive" id="output">

    class oneCluster {

        constructor(id, span) {
            this.id = id;
            this.managers = {};
            this.parentSpan = span;

            console.log("oneCluster.init div:", span);
            this.h2 = $('<h2></h2>').html(id);
            this.div = $('<div></div>').addClass("table-responsive");
            this.outputbody = $('<tbody></tbody>');
            let thead = $('<thead></thead>').html('<tr><th>#</th><th>Manager Name</th><th>Channel before</th><th>Channel after</th><th>Workers</th></tr>');
            let table = $('<table></table>').addClass("table table-striped table-sm").append(thead, this.outputbody);
            this.div.append(table);

            this.parentSpan.append(this.h2, this.div);
        }

        drawOneManager(number, data) {

            let name = data.Name;
            if (!(name in this.managers)) {
                this.managers[name] = new oneManager(number, name, this.outputbody);
            }

            let m = this.managers[name];
            m.draw(data)
        }

        drawAllManagers(data) {
            if (typeof (data) === 'string') {
                // create a JSON object
                let jsonObject = JSON.parse(data);
                if (!('ManagerData' in jsonObject)) {
                    return
                }

                for (let i = 0, len = jsonObject.ManagerData.length; i < len; i++) {
                    this.drawOneManager(i + 1, jsonObject.ManagerData[i]);
                }
            }
        }
    }

    function cleanAlertClasses(obj) {
        return obj.removeClass("alert alert-danger alert-dark alert-info alert-light" +
            "alert-primary alert-secondary alert-success alert-warning")
    }


    let ws;
    let cluster = new oneCluster('test-cluster', $('#output'));
    $('#connectionButton').click(function () {
        if (ws) {
            return false;
        }

        ws = new WebSocket(webSocketHost);
        ws.onopen = function (event) {
            $('#connectionStatus').html("Connected")
                .removeClass("alert-danger")
                .addClass('alert-success');
        }

        ws.onclose = function (event) {
            ws = null;
            cleanAlertClasses($('#connectionStatus'))
                .html("Not connected")
                .addClass('alert-danger');
        }

        ws.onmessage = function (event) {
            cleanAlertClasses($('#connectionStatus')).html("Message...").addClass('alert-info');
            cluster.drawAllManagers(event.data);
            cleanAlertClasses($('#connectionStatus')).html("Connected").addClass('alert-success');
        }

        ws.onerror = function (evt) {
            $('#connectionStatus').html("ERROR: " + evt.data)
                .removeClass("alert-success alert-info")
                .addClass('alert-danger')
        }
        return false;
    });


//    document.getElementById("send").onclick = function(evt) {
//        if (!ws) {
//            return false;
//        }
//        print("SEND: " + input.value);
//        ws.send(input.value);
//        return false;
//    };

//    document.getElementById("close").onclick = function(evt) {
//        if (!ws) {
//            return false;
//        }
//        ws.close();
//        return false;
//    };
});
