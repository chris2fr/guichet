import { client } from "$lib/pocketbase";
import type { UsersRecord } from "$lib/pocketbase/generated-types";
import type { PageLoad } from "./$types";

export const load: PageLoad = async function ({ params: { slug: id } }) {
  const user: UsersRecord =
    id === "new"
      ? ({} as UsersRecord)
      : await client.collection("users").getOne(id);

  return {
    user,
  };
};
