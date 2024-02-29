import { client } from "$lib/pocketbase";
import type { StructuresRecord, UsersRecord } from "$lib/pocketbase/generated-types";
import type { PageLoad } from "./$types";

export const load: PageLoad = async function () {
  const user: UsersRecord =
    await client.collection("users").getOne(client.authStore.model.id, {expand: 'structures'});
  return {
    user,
  };
};
