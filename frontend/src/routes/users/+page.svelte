<script lang="ts">
  import { metadata } from "$lib/app/stores";
  // import Image from "$lib/components/Image.svelte";
  import { watch } from "$lib/pocketbase";
  import type { UsersResponse } from "$lib/pocketbase/generated-types";
  $metadata.title = "Users";
  const users = watch<UsersResponse>("users", {
    sort: "-updated",
  });
</script>



<table>
  <tbody>
    {#each $users.items as user}
        <tr>
          <td>
          </td>
          <td><a href={user.username}>{user.name}</a></td>
          <td>{user.updated}</td>
          <td><a href={`${user.id}/edit`}>Edit</a></td>

        </tr>
    {:else}
      <tr>
        <td>No users found.</td>
      </tr>
    {/each}
  </tbody>
</table>
