var perso_id = 0;
var last_id = 0;

function searchDirectory() {
    var input = document.getElementById("search").value;
    if(input){
        var xhttp = new XMLHttpRequest();
        xhttp.onreadystatechange = function() {
        if (this.readyState == 4 && this.status == 201) {
        // Typical action to be performed when the document is ready:
            //Response from Request Ajax
            var jsonResponse = JSON.parse(xhttp.responseText);

            if (last_id < jsonResponse.id) {
                last_id = jsonResponse.id
                //We get the old table element, we create an new table element then we increment this new table.
                //After the new add, we replace the old table by the new one.
                var old_table = document.getElementById("users");
                var table = document.createElement('tbody');
                table.setAttribute("id","users");
            
                for (let i =0; i < Object.keys(jsonResponse.search).length; i++) {
                    var row = table.insertRow(0);
                    var urlName = row.insertCell(0);
                    var identifiant = row.insertCell(1);
                    var displayname = row.insertCell(2);
                    var email = row.insertCell(3);
                    var description = row.insertCell(4);
                    description.setAttribute("style", "word-break: break-all;");

                    if (jsonResponse.search[i].dn.localeCompare("")!=0) {
                        urlName.innerText = `<object data="/image/${jsonResponse.search[i].dn}/little" class=".img-thumbnail"><image src="/image/unknown_profile/little" class=".img-thumbnail"></object>`
                    }else {
                        urlName.innerText=""    
                    }                
                    identifiant.innerText = `<a href="/admin/ldap/${jsonResponse.search[i].dn}">${jsonResponse.search[i].id}</a>`
                    displayname.innerText = jsonResponse.search[i].displayname
                    email.innerText = jsonResponse.search[i].email
                    description.innerText = jsonResponse.search[i].description

                }
                old_table.parentNode.replaceChild(table, old_table)
            }
        }
        };
        perso_id += 1
        xhttp.overrideMimeType("application/json");
        xhttp.open("POST", "/search/".concat(input), true);
        xhttp.send(JSON.stringify({"id": perso_id}));
    } 
}