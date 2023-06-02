data "waypoint_runner_profile" "default_docker" {
  id = "01GV45AW59XGNT906S8XXKG5E5"
}

output "default_profile" {
  value = data.waypoint_runner_profile.default_docker
}
