#version 330

in vec2 base_pos;
in vec2 base_uv;

in vec2 in_origin;
in vec2 in_scale;
in vec2 in_position;
in float in_rotation;

in vec4 in_uvs;
in float in_layer;

in vec4 in_color;
in float in_additive;

uniform mat4 proj;

out vec4 col_tint;
out vec3 tex_coord;
out float additive;
void main()
{
    float cs = cos(in_rotation);
    float sn = sin(in_rotation);

    vec2 bPos = (base_pos - in_origin) * in_scale;

    vec2 sPos = vec2(cs * bPos.x - sn * bPos.y, sn * bPos.x + cs * bPos.y);

    sPos += in_position;

    gl_Position = proj * vec4(sPos, 0.0, 1.0);

    tex_coord = vec3(in_uvs[int(base_uv.x)], in_uvs[int(base_uv.y)], in_layer);
    col_tint = in_color;
    additive = in_additive;
}