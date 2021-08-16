var last_id = 0;

function searchDirectory() {
    var input = document.getElementById("search").value;
    if(input){
        last_id++;
        var request_id = last_id;

        var data = new FormData();
        data.append("query", input);

        var xhttp = new XMLHttpRequest();
        xhttp.onreadystatechange = function() {
            if (request_id != last_id) return;

            if (this.readyState == 4 && this.status == 200) {
                var result_div = document.getElementById("search-results");
                result_div.innerHTML = xhttp.responseText;
            }
        };
        xhttp.open("POST", "/directory/search", true);
        xhttp.send(data);
    } 
}
