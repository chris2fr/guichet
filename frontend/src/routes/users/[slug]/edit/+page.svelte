<script lang="ts">
  import { goto } from "$app/navigation";
  // import { authModel, save } from "$lib/pocketbase";
         import { save } from "$lib/pocketbase";
  import { alertOnFailure } from "$lib/pocketbase/ui";
  import type { PageData } from "./$types";
  export let data: PageData;
  $: ({ user } = data);
  async function submit(e: SubmitEvent) {
    // user.id = $authModel?.id;
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
  <input type="file" bind:value={user.avatar} />
  <button type="submit">Submit</button>
</form>
