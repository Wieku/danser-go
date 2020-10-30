#version 330

uniform sampler2DArray tex;
uniform vec4 col_tint;
uniform float points;
uniform float instances;

in vec2 tex_coord;
in float index;
in vec4 color_pass;

out vec4 color;

void main() {
    vec4 in_color = texture(tex, vec3(tex_coord, 0));
    color = in_color * col_tint * color_pass;// * vec4(vec3(1.0), 1.0-smoothstep(points / 3.0, points, instances - 1 - index));
    color.a *= smoothstep(instances - points, instances - points * 2 / 3, index);
}