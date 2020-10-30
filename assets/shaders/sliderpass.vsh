#version 330

in vec3 in_position;
in vec2 center;

uniform mat4 projection;
uniform float scale;
uniform mat4 distort;

void main() {
    gl_Position = projection * distort * vec4(scale * in_position.xy + center, in_position.z, 1.0);
}