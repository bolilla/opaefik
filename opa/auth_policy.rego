package system.main

default authz_public = true

default authz_authenticated_any = false
authz_authenticated_any {
  count(input.User) > 0
}

default authz_mac_vharkonen = false
authz_mac_vharkonen {
  input.User = "vharkonen"
}

default authz_group_mentat = false
authz_group_mentat {
  input.Context.group[_]="Mentat"
}

default authz_house_atreides = false
authz_house_atreides {
  input.Context.house[_]="Atreides"
  count(input.Context.house) = 1
}


