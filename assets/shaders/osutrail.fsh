#version 330

uniform sampler2DArray tex;
uniform float layer;
uniform float alpha;

in vec2 tex_coord;
in vec4 color_pass;

out vec4 color;

void main() {
    vec4 in_color = texture(tex, vec3(tex_coord, layer));

    color = in_color * color_pass;
    color.a *= alpha;
}