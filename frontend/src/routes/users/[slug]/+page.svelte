<script lang="ts">
  import { base } from "$app/paths";
  import { page } from "$app/stores";
  import { metadata } from "$lib/app/stores";
  import Delete from "$lib/components/Delete.svelte";
  import { client } from "$lib/pocketbase";
  import type { PageData } from "./$types";
  export let data: PageData;
  $: ({
    user: { id, username, email, name, avatar, structures },
  } = data);
  $: $metadata.title = name;
</script>

{#if $page.url.hash === "#delete"}
  <Delete table="users" {id} />
{/if}

{#if avatar}
  <img
    src={client.getFileUrl(data.user, avatar, { thumb: "600x0" })}
    alt={name}
  />
{/if}
<pre>{name}</pre>

<a href={`${base}/auditlog/users/${id}`}>
  <button type="button">AuditLog</button>
</a>

<style>
  img {
    max-width: 100%;
  }
</style>
