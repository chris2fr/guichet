function searchDirectory() {
    var input = document.getElementById("search").value;

    var xhttp = new XMLHttpRequest();
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4 && this.status == 201) {
        // Typical action to be performed when the document is ready:
            //Response from Request Ajax
            var jsonResponse = JSON.parse(xhttp.responseText);

            //We get the old table element, we create an new table element then we increment this new table.
            //After the new add, we replace the old table by the new one.
            var old_table = document.getElementById("users");
            var table = document.createElement('tbody');
            table.setAttribute("id","users");
            
            for (let i =0; i < Object.keys(jsonResponse).length; i++) {
                var row = table.insertRow(0);
                var identifiant = row.insertCell(0);
                var name = row.insertCell(1);
                var email = row.insertCell(2);
                identifiant.innerHTML = `<a href="/admin/ldap/${jsonResponse[i].dn}">${jsonResponse[i].identifiant}</a>`
                name.innerHTML = jsonResponse[i].name
                email.innerHTML = jsonResponse[i].email

            }
            old_table.parentNode.replaceChild(table, old_table)
        }
    };
    xhttp.overrideMimeType("application/json");
    xhttp.open("GET", "/search/".concat(input), true);
    xhttp.send();



}