<script lang="ts">
  import { goto } from "$app/navigation";
  // import { authModel, save } from "$lib/pocketbase";
         import { client, save } from "$lib/pocketbase";
    import type { NotYetMyStructuresResponse, MyStructuresResponse, StructuresResponse } from "$lib/pocketbase/generated-types";
  import { alertOnFailure } from "$lib/pocketbase/ui";
  import type { PageData } from "./$types";


  import { watch } from "$lib/pocketbase";

  const structures = watch<NotYetMyStructuresResponse>("notYetMyStructures", {
    sort: "name",
  });

  export let data: PageData;
  $: ({ user } = data);

  async function submit(e: SubmitEvent) {
    alertOnFailure(async () => {
      await save("users", user);
      goto("../..");
    });
  }

</script>

<form on:submit|preventDefault={submit}>
  <input name="name" bind:value={user.name} placeholder="name" />
  <input name="username" bind:value={user.username} placeholder="username" />
  <input name="email"  bind:value={user.email} placeholder="email" disabled />
  <input type="file" bind:files={user.avatar} />
  <button type="submit">Submit</button>
</form>

<table>
  <tbody>
    {#each user.expand.structures as structure}
        <tr>
          <td>
          </td>
          <td><a href={structure.id}>{structure.name}</a></td>
          <td><a href={`${structure.id}/removeStructure`}>Enlever</a></td>
        </tr>
    {:else}
      <tr>
        <td>Pas d'identification aux structures</td>
      </tr>
    {/each}
  </tbody>
</table>

<table>
  <tbody>
    {#each $structures.items as structure}
        <tr>
          <td>
          </td>
          <td><a href={structure.id}>{structure.name}</a></td>
          <td><a href={`${structure.id}/addStructure`}>Ajouter</a></td>
        </tr>
    {:else}
      <tr>
        <td>Pas d'autre structure pour s'y abonner</td>
      </tr>
    {/each}
  </tbody>
</table>
