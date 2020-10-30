#version 330

in vec2 in_position;
in vec2 in_tex_coord;

uniform mat4 projection;
uniform vec2 position;
uniform vec2 size;

out vec2 tex_coord;

void main() {
    gl_Position = projection * vec4(position + in_position.xy * size, 0.0, 1.0);
    tex_coord = in_tex_coord;
}