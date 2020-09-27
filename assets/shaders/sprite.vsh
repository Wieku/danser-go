#version 330

in vec2 base_pos;
in vec2 base_uv;

in vec2 in_origin;
in vec2 in_scale;
in vec2 in_position;
in float in_rotation;

in vec2 in_u;
in vec2 in_v;
in float in_layer;

in vec4 in_color;
in float in_additive;
in float in_msdf;

uniform mat4 proj;

out vec4 col_tint;
out vec3 tex_coord;
out float additive;
out float msdf;

void main()
{
    vec2 bPos = (base_pos - (in_origin * 2 - 1)) * in_scale;

    float cs = cos(in_rotation);
    float sn = sin(in_rotation);

    vec2 rPos = vec2(cs * bPos.x - sn * bPos.y, sn * bPos.x + cs * bPos.y);

    rPos += in_position;

    gl_Position = proj * vec4(rPos, 0.0, 1.0);

    tex_coord = vec3(in_u[int(base_uv.x)], in_v[int(base_uv.y)], in_layer);
    col_tint = in_color;
    additive = in_additive;
    msdf = in_msdf;
}