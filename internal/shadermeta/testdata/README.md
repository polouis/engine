# shadermeta testdata

These are the game's real compiled shaders, copied verbatim from the output of
`assetc shaders`:

    game/dist/asset/shader/PositionColor.vert.spv
    game/dist/asset/shader/SolidColor.frag.spv

Copy them from `dist/`, never from `game/asset/shaders/compiled/` -- that
directory is a leftover from before `assetc` existed and its contents do not
track the HLSL sources.

Refresh them after any change to `game/asset/shaders/source/*.hlsl` that alters
a shader's inputs or resource bindings, and update the expectations in
`shadermeta_test.go` to match. The tests exist to pin the reflector against
real bytecode, so stale copies defeat the purpose.
