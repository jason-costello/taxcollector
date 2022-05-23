<script>
        import { DataTable } from "carbon-components-svelte";

        import AutoComplete from "simple-svelte-autocomplete"
        let selectedStreet
        let selectedNeighborhood
        let properties
        let dollarUS = Intl.NumberFormat("en-US", {
                style: "currency",
                currency: "USD",
        });
        async function getStreetNames(streetName) {
                const url = "http://localhost:8777/street/"+encodeURIComponent(streetName);
                const response = await fetch(url)
                const json = await response.json()
                return json
        }

        async function getNeighborhoodNames(neighborhoodName) {
                const url = "http://localhost:8777/neighborhood/"+encodeURIComponent(neighborhoodName);
                const response = await fetch(url)
                const json = await response.json()
                console.dir(json);
                return json
        }
        let columns = []
        let values = []
        let html_rows = []
        let html_row = []
        let cols_range=[]
        let rows_range = []
        let rows_len
        let rows
        let tableData = []
        async function getPropertiesByStreetName(selectedStreet){
                const url = "http://localhost:8777/property/street/"+encodeURIComponent(selectedStreet);
                console.log("fetching: " + url)
                const response = await fetch(url)
                const json = await response.json()
                json.forEach(row => {
                        console.log(row.address)
                        console.log(row.legalDescription)
                })
                return json
        }




</script>

<h1>Property Search</h1>

<div class="todoapp stack-large">

        Street
        <AutoComplete
                searchFunction="{getStreetNames}"
                delay="200"
                localFiltering="false"
                labelFieldName="String"
                valueFieldName="String"
                bind:selectedItem="{selectedStreet}"
        />


</div>
<div>
        Neighborhood
        <AutoComplete
                searchFunction="{getNeighborhoodNames}"
                delay="200"
                localFiltering="false"
                labelFieldName="String"
                valueFieldName="String"
                bind:selectedItem="{selectedNeighborhood}"
        />

</div>

        <button on:click={getPropertiesByStreetName(selectedStreet.String)}>
                Click me
        </button>

<hr>
<div>

        <!--{#each tableData as data}-->
        <!--        <script> let data = data[data]; </script>-->
        <!--        <div class="comment" class:new={data.isNew}> ... </div>-->
        <!--{/each}-->

        <DataTable
                headers={[
    { key: "address", value: "address" },
    { key: "owners", value: "owners" },
    { key: "livableSqft", value: "livableSqft" },
    { key: "improvementAppraisal", value: "improvementAppraisal" },
    { key: "dollarPerLivableSqft", value: "dollarPerLivableSqft" },
    { key: "acerage", value: "acerage" },
    { key: "landAppraisal", value: "landAppraisal" },
    { key: "dollarPerAcre", value: "dollarPerAcre" },
    { key: "description", value: "description"}
  ]}



                rows={[

                        ]
                }

        />

</div>
